package xyz.block.ftl.runtime;

import xyz.block.ftl.Enum;

@Enum
public enum Shape {
    CIRCLE("circle"),
    SQUARE("square"),
    TRIANGLE("triangle");

    private final String value;

    Shape(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }
}
