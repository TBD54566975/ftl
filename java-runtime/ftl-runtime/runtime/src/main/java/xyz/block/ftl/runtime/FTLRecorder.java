package xyz.block.ftl.runtime;

import io.quarkus.arc.Arc;
import io.quarkus.runtime.annotations.Recorder;

@Recorder
public class FTLRecorder {

    public void registerVerb(String module, String verbName, String verbInputType, String verbOutputType, String methodName, String className) {

        //TODO: this sucks
        try {
            Class<?> verbHandlerClass = Class.forName(className, false, Thread.currentThread().getContextClassLoader());
            Class<?> inputType = Class.forName(verbInputType, false, Thread.currentThread().getContextClassLoader());
            var method = verbHandlerClass.getDeclaredMethod(methodName, inputType);
            var handlerInstance = Arc.container().instance(verbHandlerClass);
            Arc.container().instance(VerbRegistry.class).get().register(module, verbName, handlerInstance, method, inputType);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
}
