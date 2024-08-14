package xyz.block.ftl.runtime;

import jakarta.enterprise.event.Observes;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;

import io.quarkus.runtime.StartupEvent;

/**
 * This class configures the FTL serialization
 */
public class JsonSerializationConfig {

    void startup(@Observes StartupEvent event, ObjectMapper mapper) {
        mapper.configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false);
    }
}
