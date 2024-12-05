package xyz.block.ftl.runtime;

import jakarta.inject.Singleton;

import io.grpc.stub.StreamObserver;
import io.quarkus.grpc.GrpcService;
import xyz.block.ftl.language.v1.HotReloadServiceGrpc;
import xyz.block.ftl.language.v1.RunnerStartedRequest;
import xyz.block.ftl.language.v1.RunnerStartedResponse;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;
import xyz.block.ftl.v1.PingRequest;
import xyz.block.ftl.v1.PingResponse;
import xyz.block.ftl.v1.VerbServiceGrpc;

import java.util.HashMap;
import java.util.Map;

@Singleton
@GrpcService
public class HotReloadHandler extends HotReloadServiceGrpc.HotReloadServiceImplBase {

    final VerbRegistry registry;

    public HotReloadHandler(VerbRegistry registry) {
        this.registry = registry;
    }

    @Override
    public void runnerStarted(RunnerStartedRequest request, StreamObserver<RunnerStartedResponse> responseObserver) {
        Map<String, String> databases = new HashMap<>();
        for (var db : request.getDatabasesList()) {
            databases.put(db.getName(), db.getUrl());
        }

        FTLController.instance().updateRunnerConnection(request.getAddress(), databases);
    }


    @Override
    public void ping(PingRequest request, StreamObserver<PingResponse> responseObserver) {
        responseObserver.onNext(PingResponse.newBuilder().build());
        responseObserver.onCompleted();
    }
}
