package xyz.block.ftl.runtime;

import jakarta.inject.Singleton;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.Arc;

@Singleton
public class TopicHelper {

    final FTLController controller;
    final ObjectMapper mapper;

    public TopicHelper(FTLController controller, ObjectMapper mapper) {
        this.controller = controller;
        this.mapper = mapper;
    }

    public void publish(String topic, String verb, Object message) {
        try {
            controller.publishEvent(topic, verb, mapper.writeValueAsBytes(message));
        } catch (JsonProcessingException e) {
            throw new RuntimeException(e);
        }
    }

    public static TopicHelper instance() {
        return Arc.container().instance(TopicHelper.class).get();
    }
}
