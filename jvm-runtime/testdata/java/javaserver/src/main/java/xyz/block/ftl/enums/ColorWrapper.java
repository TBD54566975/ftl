package xyz.block.ftl.enums;

import org.jetbrains.annotations.NotNull;

public class ColorWrapper {
    private @NotNull ColorInt color;

    public ColorWrapper() {
    }

    public ColorWrapper(@NotNull ColorInt color) {
        this.color = color;
    }

    public ColorWrapper setColor(@NotNull ColorInt color) {
        this.color = color;
        return this;
    }

    public @NotNull ColorInt getColor() {
        return color;
    }

    public String toString() {
        return "ColorWrapper(color=" + this.color + ")";
    }
}
