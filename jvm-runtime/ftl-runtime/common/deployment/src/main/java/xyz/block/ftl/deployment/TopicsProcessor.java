package xyz.block.ftl.deployment;

import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;
import java.util.function.Consumer;

import org.jboss.jandex.DotName;
import org.jboss.jandex.Type;
import org.jboss.logging.Logger;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.deployment.GeneratedClassGizmoAdaptor;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.GeneratedClassBuildItem;
import io.quarkus.gizmo.ClassCreator;
import io.quarkus.gizmo.MethodDescriptor;
import xyz.block.ftl.ConsumableTopic;
import xyz.block.ftl.Export;
import xyz.block.ftl.WriteableTopic;
import xyz.block.ftl.runtime.TopicHelper;
import xyz.block.ftl.schema.v1.Decl;

public class TopicsProcessor {

    public static final DotName WRITEABLE_TOPIC = DotName.createSimple(WriteableTopic.class);
    public static final DotName CONSUMEABLE_TOPIC = DotName.createSimple(ConsumableTopic.class);
    private static final Logger log = Logger.getLogger(TopicsProcessor.class);

    @BuildStep
    TopicsBuildItem handleTopics(CombinedIndexBuildItem index, BuildProducer<GeneratedClassBuildItem> generatedTopicProducer,
            BuildProducer<AdditionalBeanBuildItem> beans) {
        var topicDefinitions = index.getComputingIndex().getAnnotations(FTLDotNames.TOPIC);
        log.infof("Processing %d topic definition annotations into decls", topicDefinitions.size());
        Map<DotName, TopicsBuildItem.DiscoveredTopic> topics = new HashMap<>();
        Set<String> names = new HashSet<>();
        for (var topicDefinition : topicDefinitions) {
            var iface = topicDefinition.target().asClass();
            if (!iface.isInterface()) {
                throw new RuntimeException(
                        "@TopicDefinition can only be applied to interfaces " + iface.name() + " is not an interface");
            }
            Type paramType = null;
            Type partitionMapperType = null;
            var consumer = false;
            for (var i : iface.interfaceTypes()) {
                if (i.name().equals(WRITEABLE_TOPIC)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        paramType = i.asParameterizedType().arguments().get(0);
                        partitionMapperType = i.asParameterizedType().arguments().get(1);
                    }
                } else if (i.name().equals(CONSUMEABLE_TOPIC)) {
                    consumer = true;
                }

            }
            if (paramType == null || partitionMapperType == null) {
                if (consumer) {
                    // We don't care about these here, they are handled by the subscriptions processor
                    continue;
                }
                throw new RuntimeException(
                        "@TopicDefinition can only be applied to interfaces that directly extend " + WRITEABLE_TOPIC
                                + " with a concrete type parameter " + iface.name() + " does not extend this interface");
            }
            var partitionMapperClass = partitionMapperType.name();
            beans.produce(AdditionalBeanBuildItem.unremovableOf(partitionMapperClass.toString()));

            String name = topicDefinition.value().asString();
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
                                Object.class, Class.class),
                        helper, publish.load(name), publish.readInstanceField(verb.getFieldDescriptor(), publish.getThis()),
                        publish.getMethodParam(0), publish.loadClass(partitionMapperClass.toString()));
                publish.returnVoid();
                topics.put(iface.name(), new TopicsBuildItem.DiscoveredTopic(name, cc.getClassName(), paramType,
                        iface.hasAnnotation(Export.class), iface.name().toString()));
            }
        }
        return new TopicsBuildItem(topics);
    }

    @BuildStep
    public SchemaContributorBuildItem topicSchema(TopicsBuildItem topics) {
        //register all the topics we are defining in the module definition
        return new SchemaContributorBuildItem(new Consumer<ModuleBuilder>() {
            @Override
            public void accept(ModuleBuilder moduleBuilder) {
                for (var topic : topics.getTopics().values()) {
                    moduleBuilder.addDecls(Decl.newBuilder().setTopic(xyz.block.ftl.schema.v1.Topic.newBuilder()
                            .setExport(topic.exported())
                            .setPos(PositionUtils.forClass(topic.interfaceName()))
                            .setName(topic.topicName())
                            .setEvent(moduleBuilder.buildType(topic.eventType(), topic.exported(), Nullability.NOT_NULL))
                            .build()).build());
                }
            }
        });
    }
}
