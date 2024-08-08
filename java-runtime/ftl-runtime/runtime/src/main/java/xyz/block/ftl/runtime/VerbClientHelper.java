package xyz.block.ftl.runtime;

import com.fasterxml.jackson.databind.ObjectMapper;
import io.quarkus.arc.Arc;
import jakarta.inject.Singleton;

@Singleton
public class VerbClientHelper {

    final FTLController controller;
    final ObjectMapper mapper;

    public VerbClientHelper(FTLController controller, ObjectMapper mapper) {
        this.controller = controller;
        this.mapper = mapper;
    }

    public Object call(String verb, String module, Object message, Class<?> returnType, boolean listReturnType, boolean mapReturnType) {
        try {
            var result = controller.callVerb(verb, module, mapper.writeValueAsBytes(message));
            if (listReturnType) {
                return mapper.readerForArrayOf(returnType).readValue(result);
            } else if (mapReturnType) {
                return mapper.readerForMapOf(returnType).readValue(result);
            }
            return mapper.readerFor(returnType).readValue(result);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public static VerbClientHelper instance() {
        return Arc.container().instance(VerbClientHelper.class).get();
    }
}
