package xyz.block.ftl.runtime.config;

import java.util.List;

import org.eclipse.microprofile.config.spi.ConfigSource;

import io.smallrye.config.ConfigSourceContext;
import io.smallrye.config.ConfigSourceFactory;
import xyz.block.ftl.runtime.FTLController;

public class FTLConfigSourceFactory implements ConfigSourceFactory {

    @Override
    public Iterable<ConfigSource> getConfigSources(ConfigSourceContext context) {
        var controller = FTLController.instance();
        return List.of(new FTLConfigSource(controller));
    }
}
