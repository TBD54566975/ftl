package xyz.block.ftl.runtime;

import com.fasterxml.jackson.databind.ObjectMapper;
import io.quarkus.arc.Arc;
import io.quarkus.runtime.annotations.Recorder;
import org.checkerframework.checker.units.qual.C;
import xyz.block.ftl.Topic;
import xyz.block.ftl.v1.CallRequest;

import java.lang.reflect.Type;
import java.util.List;
import java.util.function.BiFunction;

@Recorder
public class FTLRecorder {

    public void registerVerb(String module, String verbName, String methodName, List<Class<?>> parameterTypes, Class<?> verbHandlerClass,  List<BiFunction<ObjectMapper, CallRequest, Object>> paramMappers) {
        //TODO: this sucks
        try {
            var method = verbHandlerClass.getDeclaredMethod(methodName, parameterTypes.toArray(new Class[0]));
            var handlerInstance = Arc.container().instance(verbHandlerClass);
            Arc.container().instance(VerbRegistry.class).get().register(module, verbName, handlerInstance, method, paramMappers);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }


    public void registerHttpIngress(String module, String verbName) {
        try {
            Arc.container().instance(VerbRegistry.class).get().register(module, verbName, Arc.container().instance(FTLHttpHandler.class).get());
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public BiFunction<ObjectMapper, CallRequest, Object> topicSupplier(String className, String callingVerb) {
        try {
            var cls = Thread.currentThread().getContextClassLoader().loadClass(className.replace("/", "."));
            var topic = (Topic<?>) cls.getDeclaredConstructor(String.class).newInstance(callingVerb);
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
}
