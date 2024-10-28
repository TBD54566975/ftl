package xyz.block.ftl.java.test.internal;

import io.grpc.stub.StreamObserver;
import xyz.block.ftl.v1.*;

public class TestModuleServer extends ModuleServiceGrpc.ModuleServiceImplBase {
    @Override
    public void ping(PingRequest request, StreamObserver<PingResponse> responseObserver) {
        responseObserver.onNext(PingResponse.newBuilder().build());
        responseObserver.onCompleted();
    }

    @Override
    public void getModuleContext(ModuleContextRequest request, StreamObserver<ModuleContextResponse> responseObserver) {
        //TODO: add a way to test secrets and other module context values
        responseObserver.onNext(ModuleContextResponse.newBuilder().build());
    }

    @Override
    public StreamObserver<AcquireLeaseRequest> acquireLease(StreamObserver<AcquireLeaseResponse> responseObserver) {
        return super.acquireLease(responseObserver);
    }

    @Override
    public void publishEvent(PublishEventRequest request, StreamObserver<PublishEventResponse> responseObserver) {
        super.publishEvent(request, responseObserver);
    }
}
