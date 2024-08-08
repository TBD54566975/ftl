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
import xyz.block.ftl.Export;
import xyz.block.ftl.TopicDefinition;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.VerbClientDefinition;
import xyz.block.ftl.runtime.TopicHelper;
import xyz.block.ftl.runtime.VerbClientHelper;

import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

public class VerbClientsProcessor {

    public static final DotName VERB_CLIENT = DotName.createSimple(VerbClient.class);

    @BuildStep
    VerbClientBuildItem handleTopics(CombinedIndexBuildItem index, BuildProducer<GeneratedClassBuildItem> generatedClients) {
        var clientDefinitions = index.getComputingIndex().getAnnotations(VerbClientDefinition.class);
        Map<DotName, VerbClientBuildItem.DiscoveredClients> clients = new HashMap<>();
        for (var clientDefinition : clientDefinitions) {
            var iface = clientDefinition.target().asClass();
            if (!iface.isInterface()) {
                throw new RuntimeException("@VerbClientDefinition can only be applied to interfaces and " + iface.name() + " is not an interface");
            }
            Type returnType = null;
            Type paramType = null;
            for (var i : iface.interfaceTypes()) {
                if (i.name().equals(VERB_CLIENT)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        paramType = i.asParameterizedType().arguments().get(0);
                        returnType = i.asParameterizedType().arguments().get(1);
                    }
                }
            }
            if (paramType == null || returnType == null) {
                //when we don't directly extend VerbClient we just look directly for a call method
            }
            if (paramType == null || returnType == null) {
                throw new RuntimeException("@VerbClientDefinition can only be applied to interfaces that directly extend " + VERB_CLIENT + " with a concrete type parameter and " + iface.name() + " does not extend this interface");
            }

            String name = clientDefinition.value("name").asString();
            String module = clientDefinition.value("module").asString();

            try (ClassCreator cc = new ClassCreator(new GeneratedClassGizmoAdaptor(generatedClients, true), iface.name().toString() + "_fit_verbclient", null, Object.class.getName(), iface.name().toString())) {

                //TODO: map and list types
                var publish = cc.getMethodCreator("call", Object.class, Object.class);
                var helper = publish.invokeStaticMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                var results = publish.invokeVirtualMethod(MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class, String.class, Object.class, Class.class, boolean.class, boolean.class), helper, publish.load(name), publish.load(module), publish.getMethodParam(0), publish.loadClass(returnType.name().toString()), publish.load(false), publish.load(false));
                publish.returnValue(results);
                clients.put(iface.name(), new VerbClientBuildItem.DiscoveredClients(name, module,cc.getClassName()));
            }
        }
        return new VerbClientBuildItem(clients);
    }
}
