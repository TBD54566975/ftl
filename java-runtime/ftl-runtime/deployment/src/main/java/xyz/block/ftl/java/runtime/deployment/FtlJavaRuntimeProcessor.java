package xyz.block.ftl.java.runtime.deployment;

import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.FeatureBuildItem;

class FtlJavaRuntimeProcessor {

    private static final String FEATURE = "ftl-java-runtime";

    @BuildStep
    FeatureBuildItem feature() {
        return new FeatureBuildItem(FEATURE);
    }
}
