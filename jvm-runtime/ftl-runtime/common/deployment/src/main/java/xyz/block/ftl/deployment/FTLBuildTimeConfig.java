package xyz.block.ftl.deployment;

import java.util.Optional;

import io.quarkus.runtime.annotations.ConfigItem;
import io.quarkus.runtime.annotations.ConfigRoot;

@ConfigRoot(name = "ftl")
public class FTLBuildTimeConfig {

    /**
     * The FTL module name, should be set automatically during build
     */
    @ConfigItem
    public Optional<String> moduleName;
}
