package xyz.block.ftl.deployment;

import io.quarkus.deployment.GeneratedClassGizmoAdaptor;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.GeneratedClassBuildItem;
import io.quarkus.gizmo.ClassCreator;
import io.quarkus.gizmo.MethodDescriptor;
import org.jboss.jandex.DotName;
import org.jboss.jandex.Type;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.VerbClientDefinition;
import xyz.block.ftl.VerbClientEmpty;
import xyz.block.ftl.VerbClientSink;
import xyz.block.ftl.VerbClientSource;
import xyz.block.ftl.runtime.VerbClientHelper;

import java.util.HashMap;
import java.util.Map;

public class VerbClientsProcessor {

    public static final DotName VERB_CLIENT = DotName.createSimple(VerbClient.class);
    public static final DotName VERB_CLIENT_SINK = DotName.createSimple(VerbClientSink.class);
    public static final DotName VERB_CLIENT_SOURCE = DotName.createSimple(VerbClientSource.class);
    public static final DotName VERB_CLIENT_EMPTY = DotName.createSimple(VerbClientEmpty.class);

    @BuildStep
    VerbClientBuildItem handleTopics(CombinedIndexBuildItem index, BuildProducer<GeneratedClassBuildItem> generatedClients) {
        var clientDefinitions = index.getComputingIndex().getAnnotations(VerbClientDefinition.class);
        Map<DotName, VerbClientBuildItem.DiscoveredClients> clients = new HashMap<>();
        for (var clientDefinition : clientDefinitions) {
            var iface = clientDefinition.target().asClass();
            if (!iface.isInterface()) {
                throw new RuntimeException("@VerbClientDefinition can only be applied to interfaces and " + iface.name() + " is not an interface");
            }
            String name = clientDefinition.value("name").asString();
            String module = clientDefinition.value("module").asString();
            boolean found = false;
            //TODO: map and list return types
            for (var i : iface.interfaceTypes()) {
                if (i.name().equals(VERB_CLIENT)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        var paramType = i.asParameterizedType().arguments().get(0);
                        var returnType = i.asParameterizedType().arguments().get(1);
                        try (ClassCreator cc = new ClassCreator(new GeneratedClassGizmoAdaptor(generatedClients, true), iface.name().toString() + "_fit_verbclient", null, Object.class.getName(), iface.name().toString())) {
                            var publish = cc.getMethodCreator("call", Object.class, Object.class);
                            var helper = publish.invokeStaticMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            var results = publish.invokeVirtualMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class, String.class, Object.class, Class.class, boolean.class, boolean.class), helper, publish.load(name), publish.load(module), publish.getMethodParam(0), publish.loadClass(returnType.name().toString()), publish.load(false), publish.load(false));
                            publish.returnValue(results);
                            clients.put(iface.name(), new VerbClientBuildItem.DiscoveredClients(name, module,cc.getClassName()));
                        }
                        found = true;
                        break;
                    } else {
                        throw new RuntimeException("@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and " + iface.name() + " does not have concrete type parameters");
                    }
                } else if (i.name().equals(VERB_CLIENT_SINK)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        try (ClassCreator cc = new ClassCreator(new GeneratedClassGizmoAdaptor(generatedClients, true), iface.name().toString() + "_fit_verbclient", null, Object.class.getName(), iface.name().toString())) {
                            var publish = cc.getMethodCreator("call", void.class, Object.class);
                            var helper = publish.invokeStaticMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            publish.invokeVirtualMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class, String.class, Object.class, Class.class, boolean.class, boolean.class), helper, publish.load(name), publish.load(module), publish.getMethodParam(0), publish.loadClass(Void.class), publish.load(false), publish.load(false));
                            publish.returnVoid();
                            clients.put(iface.name(), new VerbClientBuildItem.DiscoveredClients(name, module,cc.getClassName()));
                        }
                        found = true;
                        break;
                    } else {
                        throw new RuntimeException("@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and " + iface.name() + " does not have concrete type parameters");
                    }
                } else if (i.name().equals(VERB_CLIENT_SOURCE)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        var returnType = i.asParameterizedType().arguments().get(0);
                        try (ClassCreator cc = new ClassCreator(new GeneratedClassGizmoAdaptor(generatedClients, true), iface.name().toString() + "_fit_verbclient", null, Object.class.getName(), iface.name().toString())) {

                            var publish = cc.getMethodCreator("call", Object.class);
                            var helper = publish.invokeStaticMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            var results = publish.invokeVirtualMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class, String.class, Object.class, Class.class, boolean.class, boolean.class), helper, publish.load(name), publish.load(module), publish.loadNull(), publish.loadClass(returnType.name().toString()), publish.load(false), publish.load(false));
                            publish.returnValue(results);
                            clients.put(iface.name(), new VerbClientBuildItem.DiscoveredClients(name, module,cc.getClassName()));
                        }
                        found = true;
                        break;
                    } else {
                        throw new RuntimeException("@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and " + iface.name() + " does not have concrete type parameters");
                    }
                } else if (i.name().equals(VERB_CLIENT_EMPTY)) {
                        try (ClassCreator cc = new ClassCreator(new GeneratedClassGizmoAdaptor(generatedClients, true), iface.name().toString() + "_fit_verbclient", null, Object.class.getName(), iface.name().toString())) {
                            var publish = cc.getMethodCreator("call", void.class);
                            var helper = publish.invokeStaticMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            publish.invokeVirtualMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class, String.class, Object.class, Class.class, boolean.class, boolean.class), helper, publish.load(name), publish.load(module), publish.loadNull(), publish.loadClass(Void.class), publish.load(false), publish.load(false));
                            publish.returnVoid();
                            clients.put(iface.name(), new VerbClientBuildItem.DiscoveredClients(name, module,cc.getClassName()));
                        }
                        found = true;
                        break;
                }
            }
            if (!found) {
                throw new RuntimeException("@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and " + iface.name() + " does not extend a verb client type");
            }


        }
        return new VerbClientBuildItem(clients);
    }
}
