package xyz.block.ftl.runtime;

import xyz.block.ftl.Enum;

@Enum
public enum ColorInt {
    RED(0),
    GREEN(1),
    BLUE(2);

    private final int value;

    ColorInt(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }
}
