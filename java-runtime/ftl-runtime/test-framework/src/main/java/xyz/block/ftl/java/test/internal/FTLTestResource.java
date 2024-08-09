package xyz.block.ftl.java.test.internal;

import io.quarkus.test.common.QuarkusTestResourceLifecycleManager;

import java.util.Map;

public class FTLTestResource implements QuarkusTestResourceLifecycleManager {

    FTLTestServer server;

    @Override
    public Map<String, String> start() {
        server = new FTLTestServer();
        server.start();
        return Map.of("ftl.endpoint", "http://localhost:" + server.getPort());
    }

    @Override
    public void stop() {
        server.stop();
    }
}
