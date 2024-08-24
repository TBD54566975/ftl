package xyz.block.ftl.java.test.internal;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.function.Function;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.ByteString;

import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import io.quarkus.arc.Arc;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;
import xyz.block.ftl.v1.PingRequest;
import xyz.block.ftl.v1.PingResponse;
import xyz.block.ftl.v1.VerbServiceGrpc;

public class TestVerbServer extends VerbServiceGrpc.VerbServiceImplBase {

    final VerbServiceGrpc.VerbServiceStub verbService;

    /**
     * TODO: this is so hacked up
     */
    static final Map<Key, Function<?, ?>> fakeVerbs = new HashMap<>();

    public TestVerbServer() {
        var channelBuilder = ManagedChannelBuilder.forAddress("127.0.0.1", 8081);
        channelBuilder.usePlaintext();
        var channel = channelBuilder.build();
        verbService = VerbServiceGrpc.newStub(channel);
    }

    @Override
    public void call(CallRequest request, StreamObserver<CallResponse> responseObserver) {
        Key key = new Key(request.getVerb().getModule(), request.getVerb().getName());
        if (fakeVerbs.containsKey(key)) {
            //TODO: YUCK YUCK YUCK
            //This all needs a refactor
            ObjectMapper mapper = Arc.container().instance(ObjectMapper.class).get();

            Function<?, ?> function = fakeVerbs.get(key);
            Class<?> type = null;
            for (var m : function.getClass().getMethods()) {
                if (m.getName().equals("apply") && m.getParameterCount() == 1) {
                    type = m.getParameterTypes()[0];
                    if (type != Object.class) {
                        break;
                    }
                }
            }
            try {
                var result = function.apply(mapper.readerFor(type).readValue(request.getBody().newInput()));
                responseObserver.onNext(
                        CallResponse.newBuilder().setBody(ByteString.copyFrom(mapper.writeValueAsBytes(result))).build());
                responseObserver.onCompleted();
            } catch (IOException e) {
                responseObserver.onError(e);
            }
            return;
        }
        verbService.call(request, responseObserver);
    }

    @Override
    public void ping(PingRequest request, StreamObserver<PingResponse> responseObserver) {
        responseObserver.onNext(PingResponse.newBuilder().build());
        responseObserver.onCompleted();
    }

    public static <P, R> void registerFakeVerb(String module, String verb, Function<P, R> verbFunction) {
        fakeVerbs.put(new Key(module, verb), verbFunction);
    }

    record Key(String module, String verb) {
    }
}
