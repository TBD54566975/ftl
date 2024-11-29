package xyz.block.ftl.deployment;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.AnnotationTarget;
import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.IndexView;
import org.jboss.jandex.MethodInfo;
import org.jboss.logging.Logger;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import xyz.block.ftl.FromOffset;
import xyz.block.ftl.Retry;
import xyz.block.ftl.schema.v1.Metadata;
import xyz.block.ftl.schema.v1.MetadataRetry;
import xyz.block.ftl.schema.v1.MetadataSubscriber;
import xyz.block.ftl.schema.v1.Ref;

public class SubscriptionProcessor {

    private static final Logger log = Logger.getLogger(SubscriptionProcessor.class);

    @BuildStep
    public void registerSubscriptions(CombinedIndexBuildItem index,
            ModuleNameBuildItem moduleNameBuildItem,
            BuildProducer<AdditionalBeanBuildItem> additionalBeanBuildItemBuildProducer,
            BuildProducer<SchemaContributorBuildItem> schemaContributorBuildItems) throws Exception {
        AdditionalBeanBuildItem.Builder beans = AdditionalBeanBuildItem.builder().setUnremovable();
        var moduleName = moduleNameBuildItem.getModuleName();
        for (var subscription : index.getIndex().getAnnotations(FTLDotNames.SUBSCRIPTION)) {
            var info = fromJandex(index.getComputingIndex(), subscription, moduleName);
            if (subscription.target().kind() != AnnotationTarget.Kind.METHOD) {
                continue;
            }
            var method = subscription.target().asMethod();
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);
            schemaContributorBuildItems.produce(generateSubscription(method, className, info));
        }
        additionalBeanBuildItemBuildProducer.produce(beans.build());
    }

    private SchemaContributorBuildItem generateSubscription(MethodInfo method, String className, SubscriptionAnnotation info) {
        return new SchemaContributorBuildItem(moduleBuilder -> {
            moduleBuilder.registerVerbMethod(method, className, false, ModuleBuilder.BodyType.REQUIRED, (builder -> {

                builder.addMetadata(Metadata.newBuilder().setSubscriber(MetadataSubscriber.newBuilder()
                        .setTopic(Ref.newBuilder().setName(info.topic()).setModule(info.module()).build())
                        .setFromOffset(
                                info.from() == FromOffset.BEGINNING
                                        ? xyz.block.ftl.schema.v1.FromOffset.FROM_OFFSET_BEGINNING
                                        : xyz.block.ftl.schema.v1.FromOffset.FROM_OFFSET_LATEST)
                        .setDeadLetter(info.deadLetter())));

                if (method.hasAnnotation(Retry.class)) {
                    RetryRecord retry = RetryRecord.fromJandex(method.annotation(Retry.class), moduleBuilder.getModuleName());

                    MetadataRetry.Builder retryBuilder = MetadataRetry.newBuilder();
                    if (!retry.catchVerb().isEmpty()) {
                        retryBuilder.setCatch(Ref.newBuilder().setModule(retry.catchModule())
                                .setName(retry.catchVerb()).build());
                    }
                    retryBuilder.setCount(retry.count())
                            .setMaxBackoff(retry.maxBackoff())
                            .setMinBackoff(retry.minBackoff());
                    builder.addMetadata(Metadata.newBuilder().setRetry(retryBuilder).build());
                }
            }));
        });
    }

    public static SubscriptionAnnotation fromJandex(IndexView indexView, AnnotationInstance subscriptions,
            String currentModuleName) {

        AnnotationValue topicClassValue = subscriptions.value("topic");
        String topicName;

        var topicClass = indexView.getClassByName(topicClassValue.asClass().name());
        AnnotationInstance annotation = topicClass.annotation(FTLDotNames.TOPIC);
        if (annotation == null) {
            throw new IllegalArgumentException(
                    "topicClass must be annotated with @TopicDefinition for subscription " + subscriptions);
        }
        topicName = annotation.value().asString();
        AnnotationValue moduleValue = annotation.value("module");
        AnnotationValue deadLetterValue = annotation.value("deadLetter");
        boolean deadLetter = deadLetterValue != null && !deadLetterValue.asString().isEmpty() && deadLetterValue.asBoolean();
        AnnotationValue from = annotation.value("from");
        FromOffset fromOffset = from == null ? FromOffset.LATEST : FromOffset.valueOf(from.asEnum());

        return new SubscriptionAnnotation(
                moduleValue == null || moduleValue.asString().isEmpty() ? currentModuleName
                        : moduleValue.asString(),
                topicName, deadLetter, fromOffset);
    }

    public record SubscriptionAnnotation(String module, String topic, boolean deadLetter, FromOffset from) {
    }
}
