package xyz.block.ftl.enums;

import xyz.block.ftl.EnumHolder;

@EnumHolder
public final class List implements ScalarOrList {
    public final java.util.List<String> value;

    public List() {
        this.value = null;
    }

    public List(java.util.List<String> value) {
        this.value = value;
    }
}
