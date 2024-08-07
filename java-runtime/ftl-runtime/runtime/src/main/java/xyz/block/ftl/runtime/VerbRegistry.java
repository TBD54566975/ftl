package xyz.block.ftl.runtime;


import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.ByteString;
import io.quarkus.arc.InstanceHandle;
import jakarta.inject.Singleton;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;

import java.io.IOException;
import java.lang.reflect.Method;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.function.BiFunction;
import java.util.function.Function;

@Singleton
public class VerbRegistry {

    final ObjectMapper mapper;


    private final Map<Key, VerbInvoker> verbs = new ConcurrentHashMap<>();


    public VerbRegistry(ObjectMapper mapper) {
        this.mapper = mapper;
    }

    public void register(String module, String name, InstanceHandle<?> verbHandlerClass, Method method, List<BiFunction<ObjectMapper, CallRequest, Object>> paramMappers) {
        verbs.put(new Key(module, name), new AnnotatedEndpointHandler(verbHandlerClass, method, paramMappers));
    }

    public void register(String module, String name, VerbInvoker verbInvoker) {
        verbs.put(new Key(module, name), verbInvoker);
    }

    public CallResponse invoke(CallRequest request) {
        VerbInvoker handler = verbs.get(new Key(request.getVerb().getModule(), request.getVerb().getName()));
        if (handler == null) {
            return CallResponse.newBuilder().setError(CallResponse.Error.newBuilder().setMessage("Verb not found").build()).build();
        }
        return handler.handle(request);
    }


    private record Key(String module, String name) {

    }


    private class AnnotatedEndpointHandler implements VerbInvoker {
        final InstanceHandle<?> verbHandlerClass;
        final Method method;
        final List<BiFunction<ObjectMapper, CallRequest, Object>> parameterSuppliers;

        private AnnotatedEndpointHandler(InstanceHandle<?> verbHandlerClass, Method method, List<BiFunction<ObjectMapper, CallRequest, Object>> parameterSuppliers) {
            this.verbHandlerClass = verbHandlerClass;
            this.method = method;
            this.parameterSuppliers = parameterSuppliers;
        }

        public CallResponse handle(CallRequest in) {
            try {
                Object[] params = new Object[parameterSuppliers.size()];
                for (int i = 0; i < parameterSuppliers.size(); i++) {
                    params[i] = parameterSuppliers.get(i).apply(mapper, in);
                }
                Object ret;
                ret = method.invoke(verbHandlerClass.get(), params);
                var mappedResponse = mapper.writer().writeValueAsBytes(ret);
                return CallResponse.newBuilder().setBody(ByteString.copyFrom(mappedResponse)).build();
            } catch (Exception e) {
                return CallResponse.newBuilder().setError(CallResponse.Error.newBuilder().setMessage(e.getMessage()).build()).build();
            }
        }
    }


    public record BodySupplier(Class<?> inputClass) implements BiFunction<ObjectMapper, CallRequest, Object> {

        @Override
        public Object apply(ObjectMapper mapper, CallRequest in) {
            try {
                return mapper.createParser(in.getBody().newInput()).readValueAs(inputClass);
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }
    }

}
