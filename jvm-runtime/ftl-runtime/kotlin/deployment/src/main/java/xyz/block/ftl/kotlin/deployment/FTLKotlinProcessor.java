package xyz.block.ftl.kotlin.deployment;

import io.quarkus.deployment.annotations.BuildStep;
import xyz.block.ftl.deployment.DefaultOptionalBuildItem;

public class FTLKotlinProcessor {
    @BuildStep
    public DefaultOptionalBuildItem defaultOptional() {
        return new DefaultOptionalBuildItem(true);
    }
}
