package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.PosixFilePermission;
import java.util.EnumSet;
import java.util.List;
import java.util.Objects;
import java.util.stream.Collectors;

import org.jboss.jandex.DotName;
import org.jboss.logging.Logger;
import org.tomlj.Toml;
import org.tomlj.TomlParseResult;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.deployment.Capabilities;
import io.quarkus.deployment.Capability;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.ApplicationArchivesBuildItem;
import io.quarkus.deployment.builditem.ApplicationInfoBuildItem;
import io.quarkus.deployment.builditem.ApplicationStartBuildItem;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.FeatureBuildItem;
import io.quarkus.deployment.builditem.LaunchModeBuildItem;
import io.quarkus.deployment.builditem.RunTimeConfigBuilderBuildItem;
import io.quarkus.deployment.builditem.ShutdownContextBuildItem;
import io.quarkus.deployment.builditem.SystemPropertyBuildItem;
import io.quarkus.deployment.builditem.nativeimage.ReflectiveClassBuildItem;
import io.quarkus.deployment.pkg.builditem.OutputTargetBuildItem;
import io.quarkus.grpc.deployment.BindableServiceBuildItem;
import io.quarkus.netty.runtime.virtual.VirtualServerChannel;
import io.quarkus.vertx.core.deployment.CoreVertxBuildItem;
import io.quarkus.vertx.core.deployment.EventLoopCountBuildItem;
import io.quarkus.vertx.http.deployment.RequireVirtualHttpBuildItem;
import io.quarkus.vertx.http.deployment.WebsocketSubProtocolsBuildItem;
import io.quarkus.vertx.http.runtime.HttpBuildTimeConfig;
import io.quarkus.vertx.http.runtime.VertxHttpRecorder;
import xyz.block.ftl.runtime.FTLDatasourceCredentials;
import xyz.block.ftl.runtime.FTLRecorder;
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

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    public void generateSchema(CombinedIndexBuildItem index,
            FTLRecorder recorder,
            OutputTargetBuildItem outputTargetBuildItem,
            ModuleNameBuildItem moduleNameBuildItem,
            TopicsBuildItem topicsBuildItem,
            VerbClientBuildItem verbClientBuildItem,
            List<SchemaContributorBuildItem> schemaContributorBuildItems) throws Exception {
        String moduleName = moduleNameBuildItem.getModuleName();

        ModuleBuilder moduleBuilder = new ModuleBuilder(index.getComputingIndex(), moduleName, topicsBuildItem.getTopics(),
                verbClientBuildItem.getVerbClients(), recorder);

        for (var i : schemaContributorBuildItems) {
            i.getSchemaContributor().accept(moduleBuilder);
        }

        Path output = outputTargetBuildItem.getOutputDirectory().resolve(SCHEMA_OUT);
        try (var out = Files.newOutputStream(output)) {
            moduleBuilder.writeTo(out);
        }

        output = outputTargetBuildItem.getOutputDirectory().resolve("main");
        try (var out = Files.newOutputStream(output)) {
            out.write(
                    """
                            #!/bin/bash
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

    /**
     * This is a huge hack that is needed until Quarkus supports both virtual and socket based HTTP
     */
    @Record(ExecutionTime.RUNTIME_INIT)
    @BuildStep
    void openSocket(ApplicationStartBuildItem start,
            LaunchModeBuildItem launchMode,
            CoreVertxBuildItem vertx,
            ShutdownContextBuildItem shutdown,
            BuildProducer<ReflectiveClassBuildItem> reflectiveClass,
            HttpBuildTimeConfig httpBuildTimeConfig,
            java.util.Optional<RequireVirtualHttpBuildItem> requireVirtual,
            EventLoopCountBuildItem eventLoopCount,
            List<WebsocketSubProtocolsBuildItem> websocketSubProtocols,
            Capabilities capabilities,
            VertxHttpRecorder recorder) throws IOException {
        reflectiveClass
                .produce(ReflectiveClassBuildItem.builder(VirtualServerChannel.class)
                        .build());
        recorder.startServer(vertx.getVertx(), shutdown,
                launchMode.getLaunchMode(), true, false,
                eventLoopCount.getEventLoopCount(),
                websocketSubProtocols.stream().map(bi -> bi.getWebsocketSubProtocols())
                        .collect(Collectors.toList()),
                launchMode.isAuxiliaryApplication(), !capabilities.isPresent(Capability.VERTX_WEBSOCKETS));
    }
}
