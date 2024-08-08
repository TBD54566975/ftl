package xyz.block.ftl.java.test.internal;

import io.grpc.stub.StreamObserver;
import xyz.block.ftl.v1.VerbServiceGrpc;
import xyz.block.ftl.v1.*;


public class TestVerbServer extends VerbServiceGrpc.VerbServiceImplBase {


    @Override
    public void call(CallRequest request, StreamObserver<CallResponse> responseObserver) {
        var response = registry.invoke(request);
        responseObserver.onNext(response);
        responseObserver.onCompleted();
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
