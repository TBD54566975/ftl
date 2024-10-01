package xyz.block.ftl.deployment;

import java.util.Collection;
import java.util.HashMap;
import java.util.Map;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.AnnotationTarget;
import org.jboss.jandex.DotName;
import org.jboss.jandex.MethodInfo;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import xyz.block.ftl.Retry;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.v1.schema.Decl;
import xyz.block.ftl.v1.schema.Metadata;
import xyz.block.ftl.v1.schema.MetadataRetry;
import xyz.block.ftl.v1.schema.MetadataSubscriber;
import xyz.block.ftl.v1.schema.Ref;

public class SubscriptionProcessor {

    private static final Logger log = LoggerFactory.getLogger(SubscriptionProcessor.class);

    @BuildStep
    SubscriptionMetaAnnotationsBuildItem subscriptionAnnotations(CombinedIndexBuildItem combinedIndexBuildItem,
            ModuleNameBuildItem moduleNameBuildItem) {
        Collection<AnnotationInstance> subscriptionAnnotations = combinedIndexBuildItem.getComputingIndex()
                .getAnnotations(Subscription.class);
        log.info("Processing {} subscription annotations into decls", subscriptionAnnotations.size());
        Map<DotName, SubscriptionMetaAnnotationsBuildItem.SubscriptionAnnotation> annotations = new HashMap<>();
        for (var subscriptions : subscriptionAnnotations) {
            if (subscriptions.target().kind() != AnnotationTarget.Kind.CLASS) {
                continue;
            }
            annotations.put(subscriptions.target().asClass().name(),
                    SubscriptionMetaAnnotationsBuildItem.fromJandex(subscriptions, moduleNameBuildItem.getModuleName()));
        }
        return new SubscriptionMetaAnnotationsBuildItem(annotations);
    }

    @BuildStep
    public void registerSubscriptions(CombinedIndexBuildItem index,
            ModuleNameBuildItem moduleNameBuildItem,
            BuildProducer<AdditionalBeanBuildItem> additionalBeanBuildItemBuildProducer,
            SubscriptionMetaAnnotationsBuildItem subscriptionMetaAnnotationsBuildItem,
            BuildProducer<SchemaContributorBuildItem> schemaContributorBuildItems) throws Exception {
        AdditionalBeanBuildItem.Builder beans = AdditionalBeanBuildItem.builder().setUnremovable();
        var moduleName = moduleNameBuildItem.getModuleName();
        for (var subscription : index.getIndex().getAnnotations(FTLDotNames.SUBSCRIPTION)) {
            var info = SubscriptionMetaAnnotationsBuildItem.fromJandex(subscription, moduleName);
            if (subscription.target().kind() != AnnotationTarget.Kind.METHOD) {
                continue;
            }
            var method = subscription.target().asMethod();
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);
            schemaContributorBuildItems.produce(generateSubscription(method, className, info));
        }
        for (var metaSub : subscriptionMetaAnnotationsBuildItem.getAnnotations().entrySet()) {
            for (var subscription : index.getIndex().getAnnotations(metaSub.getKey())) {
                if (subscription.target().kind() != AnnotationTarget.Kind.METHOD) {
                    log.warn("Subscription annotation on non-method target: {}", subscription.target());
                    continue;
                }
                var method = subscription.target().asMethod();

                String className = method.declaringClass().name().toString();
                beans.addBeanClass(className);
                schemaContributorBuildItems.produce(generateSubscription(method,
                        className,
                        metaSub.getValue()));
            }

        }
        additionalBeanBuildItemBuildProducer.produce(beans.build());
    }

    private SchemaContributorBuildItem generateSubscription(MethodInfo method, String className,
            SubscriptionMetaAnnotationsBuildItem.SubscriptionAnnotation info) {
        return new SchemaContributorBuildItem(moduleBuilder -> {

            moduleBuilder.addDecls(Decl.newBuilder()
                    .setSubscription(xyz.block.ftl.v1.schema.Subscription.newBuilder()
                            .setName(info.name())
                            .setTopic(Ref.newBuilder().setName(info.topic()).setModule(info.module()).build()))
                    .build());
            moduleBuilder.registerVerbMethod(method, className, false, ModuleBuilder.BodyType.REQUIRED, (builder -> {
                builder.addMetadata(Metadata.newBuilder().setSubscriber(MetadataSubscriber.newBuilder().setName(info.name())));
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
}
