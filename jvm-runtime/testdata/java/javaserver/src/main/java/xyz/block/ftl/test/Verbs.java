package xyz.block.ftl.test;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

import java.util.Map;

public class Verbs {

    @Export
    @Verb
    public String anyInput(JsonNode node) {
        return node.get("name").asText();
    }

    @Export
    @Verb
    public Object anyOutput(String name) {
        return Map.of("name", name);
    }


}
