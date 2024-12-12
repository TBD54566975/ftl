package xyz.block.ftl.runtime;

import org.jboss.logging.Logger;

import io.quarkus.dev.spi.HotReplacementContext;
import io.quarkus.dev.spi.HotReplacementSetup;

public class HotReloadSetup implements HotReplacementSetup {

    static volatile HotReplacementContext context;
    private static volatile String errorOutputPath;
    private static final String ERRORS_OUT = "errors.pb";

    @Override
    public void setupHotDeployment(HotReplacementContext hrc) {
        context = hrc;
    }

    static void doScan(boolean force) {
        if (context != null) {
            try {
                context.doScan(force);
            } catch (Exception e) {
                Logger.getLogger(HotReloadSetup.class).error("Failed to scan for changes", e);
            }
        }
    }
}
