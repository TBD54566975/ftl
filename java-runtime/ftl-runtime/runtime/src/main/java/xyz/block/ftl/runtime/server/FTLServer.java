package xyz.block.ftl.runtime.server;

import io.grpc.Grpc;
import io.grpc.InsecureServerCredentials;
import io.grpc.Server;
import io.grpc.protobuf.services.HealthStatusManager;
import io.grpc.protobuf.services.ProtoReflectionService;
import xyz.block.ftl.runtime.VerbHandler;

import java.io.IOException;

public class FTLServer {

    final int port;

    public FTLServer(int port) {
        this.port = port;
    }

    public void start() throws IOException {
        String hostname = null;
        HealthStatusManager health = new HealthStatusManager();
        final Server server = Grpc.newServerBuilderForPort(port, InsecureServerCredentials.create())
                .addService(new VerbHandler())
                .addService(ProtoReflectionService.newInstance())
                .addService(health.getHealthService())
                .build()
                .start();
        System.out.println("Listening on port " + port);
        Runtime.getRuntime().addShutdownHook(new Thread() {
            @Override
            public void run() {
                // Start graceful shutdown
                server.shutdown();
                try {
                    // Wait for RPCs to complete processing
                    if (!server.awaitTermination(30, TimeUnit.SECONDS)) {
                        // That was plenty of time. Let's cancel the remaining RPCs
                        server.shutdownNow();
                        // shutdownNow isn't instantaneous, so give a bit of time to clean resources up
                        // gracefully. Normally this will be well under a second.
                        server.awaitTermination(5, TimeUnit.SECONDS);
                    }
                } catch (InterruptedException ex) {
                    server.shutdownNow();
                }
            }
        });
        // This would normally be tied to the service's dependencies. For example, if HostnameGreeter
        // used a Channel to contact a required service, then when 'channel.getState() ==
        // TRANSIENT_FAILURE' we'd want to set NOT_SERVING. But HostnameGreeter has no dependencies, so
        // hard-coding SERVING is appropriate.
        health.setStatus("", ServingStatus.SERVING);
        server.awaitTermination();
    }

}
