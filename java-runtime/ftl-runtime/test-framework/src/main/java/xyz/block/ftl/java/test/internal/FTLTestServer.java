package xyz.block.ftl.java.test.internal;

import java.io.IOException;
import java.net.InetSocketAddress;

import io.grpc.Server;
import io.grpc.netty.NettyServerBuilder;

public class FTLTestServer {

    Server grpcServer;

    public void start() {

        var addr = new InetSocketAddress("127.0.0.1", 0);
        grpcServer = NettyServerBuilder.forAddress(addr)
                .addService(new TestVerbServer())
                .build();
        try {
            grpcServer.start();
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    public int getPort() {
        return grpcServer.getPort();
    }

    public void stop() {
        grpcServer.shutdown();
    }
}
