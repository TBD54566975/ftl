package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Objects;
import java.util.stream.Collectors;

import org.jboss.logging.Logger;
import org.tomlj.Toml;
import org.tomlj.TomlParseResult;

import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.ApplicationArchivesBuildItem;
import io.quarkus.deployment.builditem.ApplicationInfoBuildItem;
import io.quarkus.deployment.builditem.FeatureBuildItem;
import io.quarkus.deployment.builditem.RunTimeConfigBuilderBuildItem;
import xyz.block.ftl.runtime.config.FTLConfigSourceFactoryBuilder;

public class ModuleProcessor {

    private static final Logger log = Logger.getLogger(ModuleProcessor.class);

    private static final String FEATURE = "ftl-java-runtime";

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
    RunTimeConfigBuilderBuildItem runTimeConfigBuilderBuildItem() {
        return new RunTimeConfigBuilderBuildItem(FTLConfigSourceFactoryBuilder.class.getName());
    }

    @BuildStep
    FeatureBuildItem feature() {
        return new FeatureBuildItem(FEATURE);
    }
}
