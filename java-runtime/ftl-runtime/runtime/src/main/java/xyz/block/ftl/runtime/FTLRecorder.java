package xyz.block.ftl.runtime;

import io.quarkus.arc.Arc;
import io.quarkus.runtime.annotations.Recorder;
import org.checkerframework.checker.units.qual.C;

import java.lang.reflect.Type;

@Recorder
public class FTLRecorder {

    public void registerVerb(String module, String verbName, Class<?> inputType, String methodName, Class<?> verbHandlerClass) {
        //TODO: this sucks
        try {
            Class[] parameterTypes;
            if (inputType == void.class) {
                parameterTypes = new Class[0];
            } else {
                parameterTypes = new Class[]{inputType};
            }
            var method = verbHandlerClass.getDeclaredMethod(methodName, parameterTypes);
            var handlerInstance = Arc.container().instance(verbHandlerClass);
            Arc.container().instance(VerbRegistry.class).get().register(module, verbName, handlerInstance, method, inputType);
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
}
