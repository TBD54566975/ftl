package xyz.block.ftl.javacomments;

import xyz.block.ftl.Enum;
import xyz.block.ftl.Export;

/**
 * Comment on a value enum
 */
@Enum
@Export
public enum EnumType {
    /**
     * Comment on an enum value
     */
    PORTENTOUS("portentous");

    private final String value;

    EnumType(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }
}
