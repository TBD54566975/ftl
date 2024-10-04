package xyz.block.ftl.enums;

import org.jetbrains.annotations.NotNull;

public class ShapeWrapper {
    private @NotNull Shape shape;

    public ShapeWrapper() {
    }

    public ShapeWrapper(@NotNull Shape shape) {
        this.shape = shape;
    }

    public ShapeWrapper setShape(@NotNull Shape shape) {
        this.shape = shape;
        return this;
    }

    public @NotNull Shape getShape() {
        return shape;
    }
}
