package xyz.block.ftl.deployment;

import io.quarkus.builder.item.SimpleBuildItem;

public final class ModuleNameBuildItem extends SimpleBuildItem {

    final String moduleName;

    public ModuleNameBuildItem(String moduleName) {
        this.moduleName = moduleName;
    }

    public String getModuleName() {
        return moduleName;
    }
}
