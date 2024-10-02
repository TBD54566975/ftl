package xyz.block.ftl.deployment;

import io.quarkus.builder.item.SimpleBuildItem;

/**
 * Build item that indicates if a type with no nullability information should default to optional.
 *
 * This is different between Kotlin and Java
 */
public final class DefaultOptionalBuildItem extends SimpleBuildItem {
    final boolean defaultToOptional;

    public DefaultOptionalBuildItem(boolean defaultToOptional) {
        this.defaultToOptional = defaultToOptional;
    }

    public boolean isDefaultToOptional() {
        return defaultToOptional;
    }
}
