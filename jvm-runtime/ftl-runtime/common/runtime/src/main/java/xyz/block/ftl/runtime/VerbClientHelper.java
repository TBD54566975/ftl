package xyz.block.ftl.runtime;

import java.util.Map;

import jakarta.inject.Singleton;

import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.Arc;

@Singleton
public class VerbClientHelper {

    final ObjectMapper mapper;

    public VerbClientHelper(ObjectMapper mapper) {
        this.mapper = mapper;
    }

    public Object call(String verb, String module, Object message, Class<?> returnType, boolean listReturnType,
            boolean mapReturnType) {
        try {
            if (message == null) {
                //Unit must be an empty map
                //TODO: what about optional?
                message = Map.of();
            }
            var result = FTLController.instance().callVerb(verb, module, mapper.writeValueAsBytes(message));
            if (listReturnType) {
                return mapper.readerForArrayOf(returnType).readValue(result);
            } else if (mapReturnType) {
                return mapper.readerForMapOf(returnType).readValue(result);
            }
            if (result == null) {
                return null;
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
