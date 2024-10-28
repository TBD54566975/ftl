package xyz.block.ftl.deployment;

import java.util.Set;

import org.eclipse.microprofile.config.spi.ConfigSource;

public class StaticConfigSource implements ConfigSource {

    public static final String QUARKUS_BANNER_ENABLED = "quarkus.banner.enabled";

    @Override
    public Set<String> getPropertyNames() {
        return Set.of(QUARKUS_BANNER_ENABLED);
    }

    @Override
    public String getValue(String propertyName) {
        if (propertyName.equals(QUARKUS_BANNER_ENABLED)) {
            return "false";
        }
        return null;
    }

    @Override
    public String getName() {
        return "Quarkus Static Config Source";
    }
}
