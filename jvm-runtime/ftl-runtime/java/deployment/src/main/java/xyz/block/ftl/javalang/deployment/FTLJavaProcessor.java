package xyz.block.ftl.javalang.deployment;

import io.quarkus.deployment.annotations.BuildStep;
import xyz.block.ftl.deployment.DefaultOptionalBuildItem;

public class FTLJavaProcessor {
    @BuildStep
    public DefaultOptionalBuildItem defaultOptional() {
        return new DefaultOptionalBuildItem(false);
    }
}
