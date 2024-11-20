package xyz.block.ftl.runtime;

import io.quarkus.dev.spi.HotReplacementContext;
import io.quarkus.dev.spi.HotReplacementSetup;

public class HotReloadSetup implements HotReplacementSetup {

    static volatile HotReplacementContext context;

    @Override
    public void setupHotDeployment(HotReplacementContext hrc) {
        context = hrc;
    }

    static void doScan() {
        if (context != null) {
            try {
                context.doScan(false);
            } catch (Exception e) {
                // ignore
            }
        }
    }
}
