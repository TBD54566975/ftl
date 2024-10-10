package xyz.block.ftl.enums;

import xyz.block.ftl.EnumHolder;

@EnumHolder
public final class Scalar implements ScalarOrList {
    public final String value;

    public Scalar() {
        this.value = null;
    }

    public Scalar(String value) {
        this.value = value;
    }
}
