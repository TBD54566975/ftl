package xyz.block.ftl.runtime;

import java.lang.reflect.Constructor;
import java.lang.reflect.InvocationTargetException;
import java.util.List;
import java.util.function.BiFunction;

import org.jboss.resteasy.reactive.server.core.ResteasyReactiveRequestContext;
import org.jboss.resteasy.reactive.server.core.parameters.ParameterExtractor;

import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.Arc;
import io.quarkus.runtime.annotations.Recorder;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.v1.CallRequest;

@Recorder
public class FTLRecorder {

    public static final String X_FTL_VERB = "X-ftl-verb";

    public void registerVerb(String module, String verbName, String methodName, List<Class<?>> parameterTypes,
            Class<?> verbHandlerClass, List<BiFunction<ObjectMapper, CallRequest, Object>> paramMappers,
            boolean allowNullReturn) {
        //TODO: this sucks
        try {
            var method = verbHandlerClass.getDeclaredMethod(methodName, parameterTypes.toArray(new Class[0]));
            var handlerInstance = Arc.container().instance(verbHandlerClass);
            Arc.container().instance(VerbRegistry.class).get().register(module, verbName, handlerInstance, method,
                    paramMappers, allowNullReturn);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public void registerHttpIngress(String module, String verbName) {
        try {
            Arc.container().instance(VerbRegistry.class).get().register(module, verbName,
                    Arc.container().instance(FTLHttpHandler.class).get());
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public BiFunction<ObjectMapper, CallRequest, Object> topicSupplier(String className, String callingVerb) {
        try {
            var cls = Thread.currentThread().getContextClassLoader().loadClass(className.replace("/", "."));
            var topic = cls.getDeclaredConstructor(String.class).newInstance(callingVerb);
            return new BiFunction<ObjectMapper, CallRequest, Object>() {
                @Override
                public Object apply(ObjectMapper mapper, CallRequest callRequest) {
                    return topic;
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public BiFunction<ObjectMapper, CallRequest, Object> verbClientSupplier(String className) {
        try {
            var cls = Thread.currentThread().getContextClassLoader().loadClass(className.replace("/", "."));
            var client = cls.getDeclaredConstructor().newInstance();
            return new BiFunction<ObjectMapper, CallRequest, Object>() {
                @Override
                public Object apply(ObjectMapper mapper, CallRequest callRequest) {
                    return client;
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public BiFunction<ObjectMapper, CallRequest, Object> leaseClientSupplier() {
        return new BiFunction<ObjectMapper, CallRequest, Object>() {
            volatile LeaseClient leaseClient;

            @Override
            public Object apply(ObjectMapper mapper, CallRequest callRequest) {
                if (leaseClient == null) {
                    leaseClient = Arc.container().instance(LeaseClient.class).get();
                }
                return leaseClient;
            }
        };
    }

    public ParameterExtractor topicParamExtractor(String className) {

        try {
            var cls = Thread.currentThread().getContextClassLoader().loadClass(className.replace("/", "."));
            Constructor<?> ctor = cls.getDeclaredConstructor(String.class);
            return new ParameterExtractor() {
                @Override
                public Object extractParameter(ResteasyReactiveRequestContext context) {

                    try {
                        Object topic = ctor.newInstance(context.getHeader(X_FTL_VERB, true));
                        return topic;
                    } catch (InstantiationException | IllegalAccessException | InvocationTargetException e) {
                        throw new RuntimeException(e);
                    }
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public ParameterExtractor verbParamExtractor(String className) {
        try {
            var cls = Thread.currentThread().getContextClassLoader().loadClass(className.replace("/", "."));
            var client = cls.getDeclaredConstructor().newInstance();
            return new ParameterExtractor() {
                @Override
                public Object extractParameter(ResteasyReactiveRequestContext context) {
                    return client;
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public ParameterExtractor leaseClientExtractor() {
        try {
            return new ParameterExtractor() {

                volatile LeaseClient leaseClient;

                @Override
                public Object extractParameter(ResteasyReactiveRequestContext context) {
                    if (leaseClient == null) {
                        leaseClient = Arc.container().instance(LeaseClient.class).get();
                    }
                    return leaseClient;
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
}
