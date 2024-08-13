package xyz.block.ftl.deployment;

import java.util.HashMap;
import java.util.Map;

import jakarta.inject.Singleton;

import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.DotName;
import org.jboss.jandex.Type;

import io.quarkus.arc.deployment.GeneratedBeanBuildItem;
import io.quarkus.arc.deployment.GeneratedBeanGizmoAdaptor;
import io.quarkus.deployment.GeneratedClassGizmoAdaptor;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.GeneratedClassBuildItem;
import io.quarkus.deployment.builditem.LaunchModeBuildItem;
import io.quarkus.gizmo.ClassCreator;
import io.quarkus.gizmo.ClassOutput;
import io.quarkus.gizmo.MethodDescriptor;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.VerbClientDefinition;
import xyz.block.ftl.VerbClientEmpty;
import xyz.block.ftl.VerbClientSink;
import xyz.block.ftl.VerbClientSource;
import xyz.block.ftl.runtime.VerbClientHelper;

public class VerbClientsProcessor {

    public static final DotName VERB_CLIENT = DotName.createSimple(VerbClient.class);
    public static final DotName VERB_CLIENT_SINK = DotName.createSimple(VerbClientSink.class);
    public static final DotName VERB_CLIENT_SOURCE = DotName.createSimple(VerbClientSource.class);
    public static final DotName VERB_CLIENT_EMPTY = DotName.createSimple(VerbClientEmpty.class);
    public static final String TEST_ANNOTATION = "xyz.block.ftl.java.test.FTLManaged";

    @BuildStep
    VerbClientBuildItem handleTopics(CombinedIndexBuildItem index, BuildProducer<GeneratedClassBuildItem> generatedClients,
            BuildProducer<GeneratedBeanBuildItem> generatedBeanBuildItemBuildProducer,
            ModuleNameBuildItem moduleNameBuildItem,
            LaunchModeBuildItem launchModeBuildItem) {
        var clientDefinitions = index.getComputingIndex().getAnnotations(VerbClientDefinition.class);
        Map<DotName, VerbClientBuildItem.DiscoveredClients> clients = new HashMap<>();
        for (var clientDefinition : clientDefinitions) {
            var iface = clientDefinition.target().asClass();
            if (!iface.isInterface()) {
                throw new RuntimeException(
                        "@VerbClientDefinition can only be applied to interfaces and " + iface.name() + " is not an interface");
            }
            String name = clientDefinition.value("name").asString();
            AnnotationValue moduleValue = clientDefinition.value("module");
            String module = moduleValue == null || moduleValue.asString().isEmpty() ? moduleNameBuildItem.getModuleName()
                    : moduleValue.asString();
            boolean found = false;
            ClassOutput classOutput;
            if (launchModeBuildItem.isTest()) {
                //when running in tests we actually make these beans, so they can be injected into the tests
                //the @TestResource qualifier is used so they can only be injected into test code
                //TODO: is this the best way of handling this? revisit later

                classOutput = new GeneratedBeanGizmoAdaptor(generatedBeanBuildItemBuildProducer);
            } else {
                classOutput = new GeneratedClassGizmoAdaptor(generatedClients, true);
            }
            //TODO: map and list return types
            for (var i : iface.interfaceTypes()) {
                if (i.name().equals(VERB_CLIENT)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        var returnType = i.asParameterizedType().arguments().get(1);
                        var paramType = i.asParameterizedType().arguments().get(0);
                        try (ClassCreator cc = new ClassCreator(classOutput, iface.name().toString() + "_fit_verbclient", null,
                                Object.class.getName(), iface.name().toString())) {
                            if (launchModeBuildItem.isTest()) {
                                cc.addAnnotation(TEST_ANNOTATION);
                                cc.addAnnotation(Singleton.class);
                            }
                            var publish = cc.getMethodCreator("call", returnType.name().toString(),
                                    paramType.name().toString());
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            var results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(name), publish.load(module), publish.getMethodParam(0),
                                    publish.loadClass(returnType.name().toString()), publish.load(false), publish.load(false));
                            publish.returnValue(results);
                            publish = cc.getMethodCreator("call", Object.class, Object.class);
                            helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(name), publish.load(module), publish.getMethodParam(0),
                                    publish.loadClass(returnType.name().toString()), publish.load(false), publish.load(false));
                            publish.returnValue(results);
                            clients.put(iface.name(),
                                    new VerbClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
                        }
                        found = true;
                        break;
                    } else {
                        throw new RuntimeException(
                                "@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and "
                                        + iface.name() + " does not have concrete type parameters");
                    }
                } else if (i.name().equals(VERB_CLIENT_SINK)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        var paramType = i.asParameterizedType().arguments().get(0);
                        try (ClassCreator cc = new ClassCreator(classOutput, iface.name().toString() + "_fit_verbclient", null,
                                Object.class.getName(), iface.name().toString())) {
                            if (launchModeBuildItem.isTest()) {
                                cc.addAnnotation(TEST_ANNOTATION);
                                cc.addAnnotation(Singleton.class);
                            }
                            var publish = cc.getMethodCreator("call", void.class, paramType.name().toString());
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(name), publish.load(module), publish.getMethodParam(0),
                                    publish.loadClass(Void.class), publish.load(false), publish.load(false));
                            publish.returnVoid();
                            publish = cc.getMethodCreator("call", void.class, Object.class);
                            helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(name), publish.load(module), publish.getMethodParam(0),
                                    publish.loadClass(Void.class), publish.load(false), publish.load(false));
                            publish.returnVoid();
                            clients.put(iface.name(),
                                    new VerbClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
                        }
                        found = true;
                        break;
                    } else {
                        throw new RuntimeException(
                                "@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and "
                                        + iface.name() + " does not have concrete type parameters");
                    }
                } else if (i.name().equals(VERB_CLIENT_SOURCE)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        var returnType = i.asParameterizedType().arguments().get(0);
                        try (ClassCreator cc = new ClassCreator(classOutput, iface.name().toString() + "_fit_verbclient", null,
                                Object.class.getName(), iface.name().toString())) {
                            if (launchModeBuildItem.isTest()) {
                                cc.addAnnotation(TEST_ANNOTATION);
                                cc.addAnnotation(Singleton.class);
                            }
                            var publish = cc.getMethodCreator("call", returnType.name().toString());
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            var results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(name), publish.load(module), publish.loadNull(),
                                    publish.loadClass(returnType.name().toString()), publish.load(false), publish.load(false));
                            publish.returnValue(results);

                            publish = cc.getMethodCreator("call", Object.class);
                            helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(name), publish.load(module), publish.loadNull(),
                                    publish.loadClass(returnType.name().toString()), publish.load(false), publish.load(false));
                            publish.returnValue(results);
                            clients.put(iface.name(),
                                    new VerbClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
                        }
                        found = true;
                        break;
                    } else {
                        throw new RuntimeException(
                                "@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and "
                                        + iface.name() + " does not have concrete type parameters");
                    }
                } else if (i.name().equals(VERB_CLIENT_EMPTY)) {
                    try (ClassCreator cc = new ClassCreator(classOutput, iface.name().toString() + "_fit_verbclient", null,
                            Object.class.getName(), iface.name().toString())) {
                        if (launchModeBuildItem.isTest()) {
                            cc.addAnnotation(TEST_ANNOTATION);
                            cc.addAnnotation(Singleton.class);
                        }
                        var publish = cc.getMethodCreator("call", void.class);
                        var helper = publish.invokeStaticMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                        publish.invokeVirtualMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                        String.class, Object.class, Class.class, boolean.class, boolean.class),
                                helper, publish.load(name), publish.load(module), publish.loadNull(),
                                publish.loadClass(Void.class), publish.load(false), publish.load(false));
                        publish.returnVoid();
                        clients.put(iface.name(), new VerbClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
                    }
                    found = true;
                    break;
                }
            }
            if (!found) {
                throw new RuntimeException(
                        "@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and "
                                + iface.name() + " does not extend a verb client type");
            }
        }
        return new VerbClientBuildItem(clients);
    }
}
