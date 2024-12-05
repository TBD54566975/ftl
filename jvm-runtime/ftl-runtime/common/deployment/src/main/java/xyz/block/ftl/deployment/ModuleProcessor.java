package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.PosixFilePermission;
import java.util.Arrays;
import java.util.Base64;
import java.util.Collection;
import java.util.EnumSet;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Timer;
import java.util.TimerTask;
import java.util.stream.Collectors;

import org.jboss.jandex.DotName;
import org.jboss.logging.Logger;
import org.tomlj.Toml;
import org.tomlj.TomlParseResult;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.bootstrap.classloading.QuarkusClassLoader;
import io.quarkus.deployment.IsDevelopment;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.ApplicationArchivesBuildItem;
import io.quarkus.deployment.builditem.ApplicationInfoBuildItem;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.FeatureBuildItem;
import io.quarkus.deployment.builditem.RunTimeConfigBuilderBuildItem;
import io.quarkus.deployment.builditem.ShutdownContextBuildItem;
import io.quarkus.deployment.builditem.SystemPropertyBuildItem;
import io.quarkus.deployment.dev.RuntimeUpdatesProcessor;
import io.quarkus.deployment.pkg.builditem.OutputTargetBuildItem;
import io.quarkus.grpc.deployment.BindableServiceBuildItem;
import io.quarkus.vertx.http.deployment.RequireSocketHttpBuildItem;
import io.quarkus.vertx.http.deployment.RequireVirtualHttpBuildItem;
import xyz.block.ftl.language.v1.Error;
import xyz.block.ftl.language.v1.ErrorList;
import xyz.block.ftl.runtime.FTLDatasourceCredentials;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.HotReloadHandler;
import xyz.block.ftl.runtime.JsonSerializationConfig;
import xyz.block.ftl.runtime.TopicHelper;
import xyz.block.ftl.runtime.VerbClientHelper;
import xyz.block.ftl.runtime.VerbHandler;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.config.FTLConfigSourceFactoryBuilder;
import xyz.block.ftl.runtime.http.FTLHttpHandler;

public class ModuleProcessor {

    private static final Logger log = Logger.getLogger(ModuleProcessor.class);

    private static final String FEATURE = "ftl-java-runtime";

    private static final String SCHEMA_OUT = "schema.pb";
    private static final String ERRORS_OUT = "errors.pb";

    @BuildStep
    BindableServiceBuildItem verbService() {
        var ret = new BindableServiceBuildItem(DotName.createSimple(VerbHandler.class));
        ret.registerBlockingMethod("call");
        ret.registerBlockingMethod("publishEvent");
        ret.registerBlockingMethod("acquireLease");
        ret.registerBlockingMethod("getDeploymentContext");
        ret.registerBlockingMethod("ping");
        return ret;
    }

    @BuildStep
    BindableServiceBuildItem hotReloadService() {
        var ret = new BindableServiceBuildItem(DotName.createSimple(HotReloadHandler.class));
        ret.registerBlockingMethod("runnerStarted");
        ret.registerBlockingMethod("ping");
        return ret;
    }

    @BuildStep
    public SystemPropertyBuildItem moduleNameConfig(ApplicationInfoBuildItem applicationInfoBuildItem) {
        return new SystemPropertyBuildItem("ftl.module.name", applicationInfoBuildItem.getName());
    }

    private static volatile Timer devModeProblemTimer;

    @BuildStep(onlyIf = IsDevelopment.class)
    @Record(ExecutionTime.STATIC_INIT)
    public void reportDevModeProblems(FTLRecorder recorder, OutputTargetBuildItem outputTargetBuildItem) {
        if (devModeProblemTimer != null) {
            return;
        }
        Path errorOutput = outputTargetBuildItem.getOutputDirectory().resolve(ERRORS_OUT);
        devModeProblemTimer = new Timer("FTL Dev Mode Error Report", true);
        devModeProblemTimer.schedule(new TimerTask() {
            @Override
            public void run() {
                Throwable compileProblem = RuntimeUpdatesProcessor.INSTANCE.getCompileProblem();
                Throwable deploymentProblems = RuntimeUpdatesProcessor.INSTANCE.getDeploymentProblem();
                if (compileProblem != null || deploymentProblems != null) {
                    ErrorList.Builder builder = ErrorList.newBuilder();
                    if (compileProblem != null) {
                        builder.addErrors(Error.newBuilder()
                                .setLevel(Error.ErrorLevel.ERROR_LEVEL_ERROR)
                                .setType(Error.ErrorType.ERROR_TYPE_COMPILER)
                                .setMsg(compileProblem.getMessage())
                                .build());
                    }
                    if (deploymentProblems != null) {
                        builder.addErrors(Error.newBuilder()
                                .setLevel(Error.ErrorLevel.ERROR_LEVEL_ERROR)
                                .setType(Error.ErrorType.ERROR_TYPE_FTL)
                                .setMsg(deploymentProblems.getMessage())
                                .build());
                    }
                    try (var out = Files.newOutputStream(errorOutput)) {
                        builder.build().writeTo(out);
                    } catch (IOException e) {
                        log.error("Failed to write error list", e);
                    }
                }
            }
        }, 1000, 1000);
        ((QuarkusClassLoader) Thread.currentThread().getContextClassLoader()).addCloseTask(new Runnable() {
            @Override
            public void run() {
                devModeProblemTimer.cancel();
                devModeProblemTimer = null;
            }
        });
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
                        log.errorf("module name not found in %s", toml);
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

    /**
     * Bytecode doesn't retain comments, so they are stored in a separate file.
     */
    @BuildStep
    public CommentsBuildItem readComments() throws IOException {
        Map<String, Collection<String>> comments = new HashMap<>();
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
        return new CommentsBuildItem(comments);
    }

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    public void generateSchema(CombinedIndexBuildItem index,
            FTLRecorder recorder,
            OutputTargetBuildItem outputTargetBuildItem,
            ModuleNameBuildItem moduleNameBuildItem,
            TopicsBuildItem topicsBuildItem,
            VerbClientBuildItem verbClientBuildItem,
            DefaultOptionalBuildItem defaultOptionalBuildItem,
            List<SchemaContributorBuildItem> schemaContributorBuildItems,
            CommentsBuildItem comments) throws Exception {
        String moduleName = moduleNameBuildItem.getModuleName();

        ModuleBuilder moduleBuilder = new ModuleBuilder(index.getComputingIndex(), moduleName, topicsBuildItem.getTopics(),
                verbClientBuildItem.getVerbClients(), recorder, comments,
                defaultOptionalBuildItem.isDefaultToOptional());

        for (var i : schemaContributorBuildItems) {
            i.getSchemaContributor().accept(moduleBuilder);
        }

        log.infof("Generating module '%s' schema from %d decls", moduleName, moduleBuilder.getDeclsCount());
        Path output = outputTargetBuildItem.getOutputDirectory().resolve(SCHEMA_OUT);
        Path errorOutput = outputTargetBuildItem.getOutputDirectory().resolve(ERRORS_OUT);
        try (var out = Files.newOutputStream(output)) {
            try (var errorOut = Files.newOutputStream(errorOutput)) {
                moduleBuilder.writeTo(out, errorOut);
            }
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
                .addBeanClasses(VerbHandler.class, HotReloadHandler.class,
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

    @Record(ExecutionTime.RUNTIME_INIT)
    @BuildStep(onlyIf = IsDevelopment.class)
    void hotReload(ShutdownContextBuildItem shutdownContextBuildItem, FTLRecorder recorder) {
        recorder.startReloadTimer(shutdownContextBuildItem);
    }
}
