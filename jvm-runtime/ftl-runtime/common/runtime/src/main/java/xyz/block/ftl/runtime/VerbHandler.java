package xyz.block.ftl.runtime;

import jakarta.inject.Singleton;

import io.grpc.stub.StreamObserver;
import io.quarkus.grpc.GrpcService;
import xyz.block.ftl.v1.*;

@Singleton
@GrpcService
public class VerbHandler extends VerbServiceGrpc.VerbServiceImplBase {

    final VerbRegistry registry;

    public VerbHandler(VerbRegistry registry) {
        this.registry = registry;
    }

    @Override
    public void call(CallRequest request, StreamObserver<CallResponse> responseObserver) {
        try {
            var response = registry.invoke(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void publishEvent(PublishEventRequest request, StreamObserver<PublishEventResponse> responseObserver) {
        super.publishEvent(request, responseObserver);
    }

    @Override
    public void sendFSMEvent(SendFSMEventRequest request, StreamObserver<SendFSMEventResponse> responseObserver) {
        super.sendFSMEvent(request, responseObserver);
    }

    @Override
    public StreamObserver<AcquireLeaseRequest> acquireLease(StreamObserver<AcquireLeaseResponse> responseObserver) {
        return super.acquireLease(responseObserver);
    }

    @Override
    public void getModuleContext(ModuleContextRequest request, StreamObserver<ModuleContextResponse> responseObserver) {
        super.getModuleContext(request, responseObserver);
    }

    @Override
    public void ping(PingRequest request, StreamObserver<PingResponse> responseObserver) {
        responseObserver.onNext(PingResponse.newBuilder().build());
        responseObserver.onCompleted();
    }
}
