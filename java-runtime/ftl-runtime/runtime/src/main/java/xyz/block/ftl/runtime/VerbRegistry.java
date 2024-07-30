package xyz.block.ftl.runtime;


import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.ByteString;
import io.quarkus.arc.InstanceHandle;
import jakarta.inject.Singleton;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;

import java.lang.reflect.Method;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Singleton
public class VerbRegistry {

    final ObjectMapper mapper;


    private final Map<Key, Handler> verbs = new ConcurrentHashMap<>();

    public VerbRegistry(ObjectMapper mapper) {
        this.mapper = mapper;
    }

    public void register(String module, String name, InstanceHandle<?> verbHandlerClass, Method method, Class<?> parameterClass) {
        verbs.put(new Key(module, name), new Handler(verbHandlerClass, method, parameterClass));
    }

    public CallResponse invoke( CallRequest request) {
        Handler handler = verbs.get(new Key(request.getVerb().getModule(), request.getVerb().getName()));
        if (handler == null) {
            return CallResponse.newBuilder().setError(CallResponse.Error.newBuilder().setMessage("Verb not found").build()).build();
        }
        return handler.handle(request);
    }


    private record Key(String module, String name) {

    }

    private class Handler {
        final InstanceHandle<?> verbHandlerClass;
        final Method method;
        final Class<?> inputClass;

        private Handler(InstanceHandle<?> verbHandlerClass, Method method, Class<?> inputClass) {
            this.verbHandlerClass = verbHandlerClass;
            this.method = method;
            this.inputClass = inputClass;
        }

        public CallResponse handle(CallRequest in) {
            try {
                var body = mapper.createParser(in.getBody().newInput()).readValueAs(inputClass);
                var ret = method.invoke(verbHandlerClass.get(), body);
                var mappedResponse = mapper.writer().writeValueAsBytes(ret);
                return CallResponse.newBuilder().setBody(ByteString.copyFrom(mappedResponse)).build();
            } catch (Exception e) {
                return CallResponse.newBuilder().setError(CallResponse.Error.newBuilder().setMessage(e.getMessage()).build()).build();
            }
        }
    }
}
