package xyz.block.ftl.enums;

import java.util.Map;

import com.fasterxml.jackson.databind.JsonNode;

import ftl.javaserver.Shape;
import ftl.javaserver.StringEnumVerbClient;
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
    public ColorWrapper valueEnumVerb(ColorWrapper color) {
        return color;
    }

    @Export
    @Verb
    public ShapeWrapper stringEnumVerb(ShapeWrapper shape) {
        return shape;
    }

    @Export
    @Verb
    public AnimalWrapper typeEnumVerb(AnimalWrapper animal) {
        return animal;
    }

    @Export
    @Verb
    public ftl.javaserver.ShapeWrapper localVerbCall(StringEnumVerbClient client) {
        return client.stringEnumVerb(new ftl.javaserver.ShapeWrapper(Shape.SQUARE));
    }
}
