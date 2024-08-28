package xyz.block.ftl.deployment;

import org.jboss.jandex.Type;

import io.quarkus.builder.item.MultiBuildItem;

public final class TypeAliasBuildItem extends MultiBuildItem {

    final String name;
    final String module;
    final Type localType;
    final Type serializedType;
    final boolean exported;

    public TypeAliasBuildItem(String name, String module, Type localType, Type serializedType, boolean exported) {
        this.name = name;
        this.module = module;
        this.localType = localType;
        this.serializedType = serializedType;
        this.exported = exported;
    }

    public String getName() {
        return name;
    }

    public String getModule() {
        return module;
    }

    public Type getLocalType() {
        return localType;
    }

    public Type getSerializedType() {
        return serializedType;
    }

    public boolean isExported() {
        return exported;
    }
}
