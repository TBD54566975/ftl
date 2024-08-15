package xyz.block.ftl.deployment;

import org.jboss.jandex.AnnotationInstance;

public record RetryRecord(int count, String minBackoff, String maxBackoff, String catchModule, String catchVerb) {

    public static RetryRecord fromJandex(AnnotationInstance nested, String currentModuleName) {
        return new RetryRecord(
                nested.value("count") != null ? nested.value("count").asInt() : 0,
                nested.value("minBackoff") != null ? nested.value("minBackoff").asString() : "",
                nested.value("maxBackoff") != null ? nested.value("maxBackoff").asString() : "",
                nested.value("catchModule") != null ? nested.value("catchModule").asString() : currentModuleName,
                nested.value("catchVerb") != null ? nested.value("catchVerb").asString() : "");
    }

    public boolean isEmpty() {
        return count == 0 && minBackoff.isEmpty() && maxBackoff.isEmpty() && catchVerb.isEmpty();
    }
}
