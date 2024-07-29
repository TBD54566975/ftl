package xyz.block.ftl.runtime;

import io.grpc.stub.StreamObserver;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;
import xyz.block.ftl.v1.VerbServiceGrpc;

public class VerbHandler extends VerbServiceGrpc.VerbServiceImplBase {

    @Override
    public void call(CallRequest request, StreamObserver<CallResponse> responseObserver) {
        super.call(request, responseObserver);
    }
}
