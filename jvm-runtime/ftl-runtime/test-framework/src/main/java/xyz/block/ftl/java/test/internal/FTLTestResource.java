package xyz.block.ftl.java.test.internal;

import java.util.Map;

import io.quarkus.test.common.QuarkusTestResourceLifecycleManager;

public class FTLTestResource implements QuarkusTestResourceLifecycleManager {

    FTLTestServer server;

    @Override
    public Map<String, String> start() {
        server = new FTLTestServer();
        server.start();
        String endpoint = "http://127.0.0.1:" + server.getPort();
        System.setProperty("ftl.test.endpoint", endpoint);
        return Map.of("ftl.endpoint", endpoint);
    }

    @Override
    public void stop() {
        server.stop();
    }

    @Override
    public void inject(TestInjector testInjector) {

    }
}
