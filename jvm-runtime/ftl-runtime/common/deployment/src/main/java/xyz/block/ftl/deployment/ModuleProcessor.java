package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.PosixFilePermission;
import java.util.Arrays;
import java.util.Base64;
import java.util.EnumSet;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.stream.Collectors;

import org.jboss.jandex.DotName;
import org.jboss.jandex.ParameterizedType;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.tomlj.Toml;
import org.tomlj.TomlParseResult;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.ApplicationArchivesBuildItem;
import io.quarkus.deployment.builditem.ApplicationInfoBuildItem;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.FeatureBuildItem;
import io.quarkus.deployment.builditem.RunTimeConfigBuilderBuildItem;
import io.quarkus.deployment.builditem.SystemPropertyBuildItem;
import io.quarkus.deployment.pkg.builditem.OutputTargetBuildItem;
import io.quarkus.grpc.deployment.BindableServiceBuildItem;
import io.quarkus.vertx.http.deployment.RequireSocketHttpBuildItem;
import io.quarkus.vertx.http.deployment.RequireVirtualHttpBuildItem;
import xyz.block.ftl.runtime.FTLDatasourceCredentials;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.JsonSerializationConfig;
import xyz.block.ftl.runtime.TopicHelper;
import xyz.block.ftl.runtime.VerbClientHelper;
import xyz.block.ftl.runtime.VerbHandler;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.config.FTLConfigSourceFactoryBuilder;
import xyz.block.ftl.runtime.http.FTLHttpHandler;
import xyz.block.ftl.v1.schema.Ref;

public class ModuleProcessor {

    private static final Logger log = LoggerFactory.getLogger(ModuleProcessor.class);

    private static final String FEATURE = "ftl-java-runtime";

    private static final String SCHEMA_OUT = "schema.pb";

    @BuildStep
    BindableServiceBuildItem verbService() {
        var ret = new BindableServiceBuildItem(DotName.createSimple(VerbHandler.class));
        ret.registerBlockingMethod("call");
        ret.registerBlockingMethod("publishEvent");
        ret.registerBlockingMethod("sendFSMEvent");
        ret.registerBlockingMethod("acquireLease");
        ret.registerBlockingMethod("getModuleContext");
        ret.registerBlockingMethod("ping");
        return ret;
    }

    @BuildStep
    public SystemPropertyBuildItem moduleNameConfig(ApplicationInfoBuildItem applicationInfoBuildItem) {
        return new SystemPropertyBuildItem("ftl.module.name", applicationInfoBuildItem.getName());
    }

    @BuildStep
    ModuleNameBuildItem moduleName(ApplicationInfoBuildItem applicationInfoBuildItem,
            ApplicationArchivesBuildItem archivesBuildItem) throws IOException {
        for (var root : archivesBuildItem.getRootArchive().getRootDirectories()) {
            Path source = root;
            for (;;) {
                var toml = source.resolve("ftl.toml");
                if (Files.exists(toml)) {
                    TomlParseResult result = Toml.parse(toml);
                    if (result.hasErrors()) {
                        throw new RuntimeException("Failed to parse " + toml + " "
                                + result.errors().stream().map(Objects::toString).collect(Collectors.joining(", ")));
                    }

                    String value = result.getString("module");
                    if (value != null) {
                        return new ModuleNameBuildItem(value);
                    } else {
                        log.error("module name not found in {}", toml);
                    }
                }
                if (source.getParent() == null) {
                    break;
                } else {
                    source = source.getParent();
                }
            }
        }

        return new ModuleNameBuildItem(applicationInfoBuildItem.getName());
    }

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    public void generateSchema(CombinedIndexBuildItem index,
            FTLRecorder recorder,
            OutputTargetBuildItem outputTargetBuildItem,
            ModuleNameBuildItem moduleNameBuildItem,
            TopicsBuildItem topicsBuildItem,
            VerbClientBuildItem verbClientBuildItem,
            List<TypeAliasBuildItem> typeAliasBuildItems,
            List<SchemaContributorBuildItem> schemaContributorBuildItems) throws Exception {
        log.info("Generating module '{}' schema from build items", moduleNameBuildItem.getModuleName());
        String moduleName = moduleNameBuildItem.getModuleName();
        Map<String, Iterable<String>> comments = readComments();
        Map<TypeKey, ModuleBuilder.ExistingRef> existingRefs = new HashMap<>();
        for (var i : typeAliasBuildItems) {
            String mn;
            if (i.getModule().isEmpty()) {
                mn = moduleNameBuildItem.getModuleName();
            } else {
                mn = i.getModule();
            }
            if (i.getLocalType() instanceof ParameterizedType) {
                //TODO: we can't handle this yet
                // existingRefs.put(new TypeKey(i.getLocalType().name().toString(), i.getLocalType().asParameterizedType().arguments().stream().map(i.)), new ModuleBuilder.ExistingRef(Ref.newBuilder().setModule(moduleName).setName(i.getName()).build(), i.isExported()));
            } else {
                existingRefs.put(new TypeKey(i.getLocalType().name().toString(), List.of()), new ModuleBuilder.ExistingRef(
                        Ref.newBuilder().setModule(mn).setName(i.getName()).build(), i.isExported()));
            }
        }

        ModuleBuilder moduleBuilder = new ModuleBuilder(index.getComputingIndex(), moduleName, topicsBuildItem.getTopics(),
                verbClientBuildItem.getVerbClients(), recorder, comments, existingRefs);

        for (var i : schemaContributorBuildItems) {
            i.getSchemaContributor().accept(moduleBuilder);
        }

        Path output = outputTargetBuildItem.getOutputDirectory().resolve(SCHEMA_OUT);
        try (var out = Files.newOutputStream(output)) {
            moduleBuilder.writeTo(out);
        }

        output = outputTargetBuildItem.getOutputDirectory().resolve("launch");
        try (var out = Files.newOutputStream(output)) {
            out.write(
                    """
                            #!/bin/bash
                            if [ -n "$FTL_DEBUG_PORT" ]; then
                                FTL_JVM_OPTS="$FTL_JVM_OPTS -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:$FTL_DEBUG_PORT"
                            fi
                            exec java $FTL_JVM_OPTS -jar quarkus-app/quarkus-run.jar"""
                            .getBytes(StandardCharsets.UTF_8));
        }
        var perms = Files.getPosixFilePermissions(output);
        EnumSet<PosixFilePermission> newPerms = EnumSet.copyOf(perms);
        newPerms.add(PosixFilePermission.GROUP_EXECUTE);
        newPerms.add(PosixFilePermission.OWNER_EXECUTE);
        Files.setPosixFilePermissions(output, newPerms);
    }

    @BuildStep
    RunTimeConfigBuilderBuildItem runTimeConfigBuilderBuildItem() {
        return new RunTimeConfigBuilderBuildItem(FTLConfigSourceFactoryBuilder.class.getName());
    }

    @BuildStep
    FeatureBuildItem feature() {
        return new FeatureBuildItem(FEATURE);
    }

    @BuildStep
    AdditionalBeanBuildItem beans() {
        return AdditionalBeanBuildItem.builder()
                .addBeanClasses(VerbHandler.class,
                        VerbRegistry.class, FTLHttpHandler.class,
                        TopicHelper.class, VerbClientHelper.class, JsonSerializationConfig.class,
                        FTLDatasourceCredentials.class)
                .setUnremovable().build();
    }

    @BuildStep
    void openSocket(BuildProducer<RequireVirtualHttpBuildItem> virtual,
            BuildProducer<RequireSocketHttpBuildItem> socket) throws IOException {
        socket.produce(RequireSocketHttpBuildItem.MARKER);
        virtual.produce(RequireVirtualHttpBuildItem.MARKER);
    }

    /**
     * Bytecode doesn't retain comments, so they are stored in a separate file
     * Each line is a key value pair separated by an '='. The key is the DeclRef and the value is the comment
     */
    private Map<String, Iterable<String>> readComments() throws IOException {
        Map<String, Iterable<String>> comments = new HashMap<>();
        try (var input = Thread.currentThread().getContextClassLoader().getResourceAsStream("META-INF/ftl-verbs.txt")) {
            if (input != null) {
                var contents = new String(input.readAllBytes(), StandardCharsets.UTF_8).split("\n");
                for (var content : contents) {
                    var eq = content.indexOf('=');
                    if (eq == -1) {
                        continue;
                    }
                    String key = content.substring(0, eq);
                    String value = new String(Base64.getDecoder().decode(content.substring(eq + 1)), StandardCharsets.UTF_8);
                    comments.put(key, Arrays.asList(value.split("\n")));
                }
            }
        }
        return comments;
    }
}
