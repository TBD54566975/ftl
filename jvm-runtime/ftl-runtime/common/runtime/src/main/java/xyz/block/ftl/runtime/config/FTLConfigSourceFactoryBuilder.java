package xyz.block.ftl.runtime.config;

import io.quarkus.runtime.configuration.ConfigBuilder;
import io.smallrye.config.SmallRyeConfigBuilder;

public class FTLConfigSourceFactoryBuilder implements ConfigBuilder {
    @Override
    public SmallRyeConfigBuilder configBuilder(SmallRyeConfigBuilder builder) {
        builder.withSources(new FTLConfigSourceFactory());
        return builder;
    }
}
