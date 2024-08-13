package xyz.block.ftl.deployment;

import java.util.Map;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.DotName;

import io.quarkus.builder.item.SimpleBuildItem;

public final class SubscriptionMetaAnnotationsBuildItem extends SimpleBuildItem {

    private final Map<DotName, SubscriptionAnnotation> annotations;

    public SubscriptionMetaAnnotationsBuildItem(Map<DotName, SubscriptionAnnotation> annotations) {
        this.annotations = annotations;
    }

    public Map<DotName, SubscriptionAnnotation> getAnnotations() {
        return annotations;
    }

    public record SubscriptionAnnotation(String module, String topic, String name) {
    }

    public static SubscriptionAnnotation fromJandex(AnnotationInstance subscriptions, String currentModuleName) {
        AnnotationValue moduleValue = subscriptions.value("module");

        return new SubscriptionMetaAnnotationsBuildItem.SubscriptionAnnotation(
                moduleValue == null || moduleValue.asString().isEmpty() ? currentModuleName
                        : moduleValue.asString(),
                subscriptions.value("topic").asString(),
                subscriptions.value("name").asString());
    }
}
