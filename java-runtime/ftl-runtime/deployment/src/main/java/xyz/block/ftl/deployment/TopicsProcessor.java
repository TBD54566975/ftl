package xyz.block.ftl.deployment;

import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

import org.jboss.jandex.AnnotationTarget;
import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.DotName;
import org.jboss.jandex.Type;

import io.quarkus.deployment.GeneratedClassGizmoAdaptor;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.GeneratedClassBuildItem;
import io.quarkus.gizmo.ClassCreator;
import io.quarkus.gizmo.MethodDescriptor;
import xyz.block.ftl.Export;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.Topic;
import xyz.block.ftl.TopicDefinition;
import xyz.block.ftl.runtime.TopicHelper;

public class TopicsProcessor {

    public static final DotName TOPIC = DotName.createSimple(Topic.class);

    @BuildStep
    TopicsBuildItem handleTopics(CombinedIndexBuildItem index, BuildProducer<GeneratedClassBuildItem> generatedTopicProducer) {
        var topicDefinitions = index.getComputingIndex().getAnnotations(TopicDefinition.class);
        Map<DotName, TopicsBuildItem.DiscoveredTopic> topics = new HashMap<>();
        Set<String> names = new HashSet<>();
        for (var topicDefinition : topicDefinitions) {
            var iface = topicDefinition.target().asClass();
            if (!iface.isInterface()) {
                throw new RuntimeException(
                        "@TopicDefinition can only be applied to interfaces " + iface.name() + " is not an interface");
            }
            Type paramType = null;
            for (var i : iface.interfaceTypes()) {
                if (i.name().equals(TOPIC)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        paramType = i.asParameterizedType().arguments().get(0);
                    }
                }

            }
            if (paramType == null) {
                throw new RuntimeException("@TopicDefinition can only be applied to interfaces that directly extend " + TOPIC
                        + " with a concrete type parameter " + iface.name() + " does not extend this interface");
            }

            String name = topicDefinition.value("name").asString();
            if (names.contains(name)) {
                throw new RuntimeException("Multiple topic definitions found for topic " + name);
            }
            names.add(name);
            try (ClassCreator cc = new ClassCreator(new GeneratedClassGizmoAdaptor(generatedTopicProducer, true),
                    iface.name().toString() + "_fit_topic", null, Object.class.getName(), iface.name().toString())) {
                var verb = cc.getFieldCreator("verb", String.class);
                var constructor = cc.getConstructorCreator(String.class);
                constructor.invokeSpecialMethod(MethodDescriptor.ofMethod(Object.class, "<init>", void.class),
                        constructor.getThis());
                constructor.writeInstanceField(verb.getFieldDescriptor(), constructor.getThis(), constructor.getMethodParam(0));
                constructor.returnVoid();
                var publish = cc.getMethodCreator("publish", void.class, Object.class);
                var helper = publish
                        .invokeStaticMethod(MethodDescriptor.ofMethod(TopicHelper.class, "instance", TopicHelper.class));
                publish.invokeVirtualMethod(
                        MethodDescriptor.ofMethod(TopicHelper.class, "publish", void.class, String.class, String.class,
                                Object.class),
                        helper, publish.load(name), publish.readInstanceField(verb.getFieldDescriptor(), publish.getThis()),
                        publish.getMethodParam(0));
                publish.returnVoid();
                topics.put(iface.name(), new TopicsBuildItem.DiscoveredTopic(name, cc.getClassName(), paramType,
                        iface.hasAnnotation(Export.class)));
            }
        }
        return new TopicsBuildItem(topics);
    }

    @BuildStep
    SubscriptionMetaAnnotationsBuildItem subscriptionAnnotations(CombinedIndexBuildItem combinedIndexBuildItem,
            ModuleNameBuildItem moduleNameBuildItem) {
        Map<DotName, SubscriptionMetaAnnotationsBuildItem.SubscriptionAnnotation> annotations = new HashMap<>();
        for (var subscriptions : combinedIndexBuildItem.getComputingIndex().getAnnotations(Subscription.class)) {
            if (subscriptions.target().kind() != AnnotationTarget.Kind.TYPE) {
                continue;
            }
            AnnotationValue moduleValue = subscriptions.value("module");
            annotations.put(subscriptions.target().asClass().name(),
                    new SubscriptionMetaAnnotationsBuildItem.SubscriptionAnnotation(
                            moduleValue == null || moduleValue.asString().isEmpty() ? moduleNameBuildItem.getModuleName()
                                    : moduleValue.asString(),
                            subscriptions.value("topic").asString(),
                            subscriptions.value("name").asString()));
        }
        return new SubscriptionMetaAnnotationsBuildItem(annotations);
    }
}
