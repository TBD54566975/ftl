package xyz.block.ftl.deployment;

import java.util.Map;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.DotName;
import org.jboss.jandex.IndexView;

import io.quarkus.builder.item.SimpleBuildItem;
import xyz.block.ftl.FromOffset;
import xyz.block.ftl.WriteableTopic;

public final class SubscriptionMetaAnnotationsBuildItem extends SimpleBuildItem {

    private final Map<DotName, SubscriptionAnnotation> annotations;

    public SubscriptionMetaAnnotationsBuildItem(Map<DotName, SubscriptionAnnotation> annotations) {
        this.annotations = annotations;
    }

    public Map<DotName, SubscriptionAnnotation> getAnnotations() {
        return annotations;
    }

    public record SubscriptionAnnotation(String module, String topic) {
    }

    public static SubscriptionAnnotation fromJandex(IndexView indexView, AnnotationInstance subscriptions,
            String currentModuleName) {

        AnnotationValue moduleValue = subscriptions.value("module");
        AnnotationValue topicValue = subscriptions.value("topic");
        AnnotationValue topicClassValue = subscriptions.value("topicClass");
        String topicName;
        if (topicValue != null && !topicValue.asString().isEmpty()) {
            if (topicClassValue != null && !topicClassValue.asClass().name().toString().equals(WriteableTopic.class.getName())) {
                throw new IllegalArgumentException("Cannot specify both topic and topicClass");
            }
            topicName = topicValue.asString();
        } else if (topicClassValue != null && !topicClassValue.asClass().name().toString().equals(WriteableTopic.class.getName())) {
            var topicClass = indexView.getClassByName(topicClassValue.asClass().name());
            AnnotationInstance annotation = topicClass.annotation(FTLDotNames.TOPIC_DEFINITION);
            if (annotation == null) {
                throw new IllegalArgumentException(
                        "topicClass must be annotated with @TopicDefinition for subscription " + subscriptions);
            }
            topicName = annotation.value().asString();
        } else {
            throw new IllegalArgumentException("Either topic or topicClass must be specified on " + subscriptions);
        }
        return new SubscriptionMetaAnnotationsBuildItem.SubscriptionAnnotation(
                moduleValue == null || moduleValue.asString().isEmpty() ? currentModuleName
                        : moduleValue.asString(),
                topicName);
    }

    public record SubscriptionOptionsAnnotation(FromOffset from, boolean deadLetter) {
        public static SubscriptionOptionsAnnotation fromJandex(AnnotationInstance subscriptionsOptions) {
            AnnotationValue deadLetterValue = subscriptionsOptions.value("deadLetter");
            return new SubscriptionOptionsAnnotation(
                    FromOffset.valueOf(subscriptionsOptions.value("from").asEnum()),
                    deadLetterValue == null || deadLetterValue.asString().isEmpty() ? false : deadLetterValue.asBoolean());
        }
    }
}
