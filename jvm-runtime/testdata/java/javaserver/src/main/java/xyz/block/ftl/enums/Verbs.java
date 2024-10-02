package xyz.block.ftl.enums;

import java.util.Map;

import com.fasterxml.jackson.databind.JsonNode;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

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

    @Export
    @Verb
    public ColorInt valueEnumVerb(ColorInt color) {
        return color;
    }

    @Export
    @Verb
    public Shape stringEnumVerb(Shape shape) {
        return shape;
    }

    @Export
    @Verb
    public Animal typeEnumVerb(Animal animal) {
        return animal;
    }
}
