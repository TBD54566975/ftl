package xyz.block.ftl.runtime;

import java.util.HashMap;
import java.util.Map;

import jakarta.inject.Singleton;

import org.jboss.logging.Logger;

import io.grpc.stub.StreamObserver;
import io.quarkus.grpc.GrpcService;
import xyz.block.ftl.language.v1.HotReloadServiceGrpc;
import xyz.block.ftl.language.v1.RunnerStartedRequest;
import xyz.block.ftl.language.v1.RunnerStartedResponse;
import xyz.block.ftl.v1.PingRequest;
import xyz.block.ftl.v1.PingResponse;

@Singleton
@GrpcService
public class HotReloadHandler extends HotReloadServiceGrpc.HotReloadServiceImplBase {
    private static final Logger log = Logger.getLogger(HotReloadHandler.class);

    @Override
    public void runnerStarted(RunnerStartedRequest request, StreamObserver<RunnerStartedResponse> responseObserver) {
        Map<String, String> databases = new HashMap<>();
        for (var db : request.getDatabasesList()) {
            databases.put(db.getName(), db.getUrl());
        }

        FTLController.instance().updateRunnerConnection(request.getAddress(), databases);
        responseObserver.onNext(RunnerStartedResponse.newBuilder().build());
        responseObserver.onCompleted();
    }

    @Override
    public void ping(PingRequest request, StreamObserver<PingResponse> responseObserver) {
        responseObserver.onNext(PingResponse.newBuilder().build());
        responseObserver.onCompleted();
    }
}
