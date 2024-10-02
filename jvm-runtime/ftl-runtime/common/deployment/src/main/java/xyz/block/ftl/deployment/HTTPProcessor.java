package xyz.block.ftl.deployment;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.function.Consumer;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.ArrayType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.PrimitiveType;
import org.jboss.jandex.VoidType;
import org.jboss.resteasy.reactive.common.model.MethodParameter;
import org.jboss.resteasy.reactive.common.model.ParameterType;
import org.jboss.resteasy.reactive.server.core.parameters.ParameterExtractor;
import org.jboss.resteasy.reactive.server.mapping.URITemplate;
import org.jboss.resteasy.reactive.server.processor.scanning.MethodScanner;

import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.resteasy.reactive.server.deployment.ResteasyReactiveResourceMethodEntriesBuildItem;
import io.quarkus.resteasy.reactive.server.spi.MethodScannerBuildItem;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.builtin.HttpRequest;
import xyz.block.ftl.runtime.builtin.HttpResponse;
import xyz.block.ftl.v1.schema.Array;
import xyz.block.ftl.v1.schema.Decl;
import xyz.block.ftl.v1.schema.IngressPathComponent;
import xyz.block.ftl.v1.schema.IngressPathLiteral;
import xyz.block.ftl.v1.schema.IngressPathParameter;
import xyz.block.ftl.v1.schema.Metadata;
import xyz.block.ftl.v1.schema.MetadataIngress;
import xyz.block.ftl.v1.schema.Ref;
import xyz.block.ftl.v1.schema.Type;
import xyz.block.ftl.v1.schema.Unit;

public class HTTPProcessor {

    @BuildStep
    @Record(ExecutionTime.STATIC_INIT)
    public MethodScannerBuildItem methodScanners(TopicsBuildItem topics,
            VerbClientBuildItem verbClients, FTLRecorder recorder) {
        return new MethodScannerBuildItem(new MethodScanner() {
            @Override
            public ParameterExtractor handleCustomParameter(org.jboss.jandex.Type type,
                    Map<DotName, AnnotationInstance> annotations, boolean field, Map<String, Object> methodContext) {
                try {

                    if (annotations.containsKey(FTLDotNames.SECRET)) {
                        Class<?> paramType = ModuleBuilder.loadClass(type);
                        String name = annotations.get(FTLDotNames.SECRET).value().asString();
                        return new VerbRegistry.SecretSupplier(name, paramType);
                    } else if (annotations.containsKey(FTLDotNames.CONFIG)) {
                        Class<?> paramType = ModuleBuilder.loadClass(type);
                        String name = annotations.get(FTLDotNames.CONFIG).value().asString();
                        return new VerbRegistry.ConfigSupplier(name, paramType);
                    } else if (topics.getTopics().containsKey(type.name())) {
                        var topic = topics.getTopics().get(type.name());
                        return recorder.topicParamExtractor(topic.generatedProducer());
                    } else if (verbClients.getVerbClients().containsKey(type.name())) {
                        var client = verbClients.getVerbClients().get(type.name());
                        return recorder.verbParamExtractor(client.generatedClient());
                    } else if (FTLDotNames.LEASE_CLIENT.equals(type.name())) {
                        return recorder.leaseClientExtractor();
                    }
                    return null;
                } catch (ClassNotFoundException e) {
                    throw new RuntimeException(e);
                }
            }
        });
    }

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    public SchemaContributorBuildItem registerHttpHandlers(
            FTLRecorder recorder,
            ResteasyReactiveResourceMethodEntriesBuildItem restEndpoints) {
        return new SchemaContributorBuildItem(new Consumer<ModuleBuilder>() {
            @Override
            public void accept(ModuleBuilder moduleBuilder) {
                //TODO: make this composable so it is not just one big method, build items should contribute to the schema
                for (var endpoint : restEndpoints.getEntries()) {
                    //TODO: naming
                    var verbName = ModuleBuilder.methodToName(endpoint.getMethodInfo());
                    boolean base64 = false;

                    //TODO: handle type parameters properly
                    org.jboss.jandex.Type bodyParamType = VoidType.VOID;
                    MethodParameter[] parameters = endpoint.getResourceMethod().getParameters();
                    for (int i = 0, parametersLength = parameters.length; i < parametersLength; i++) {
                        var param = parameters[i];
                        if (param.parameterType.equals(ParameterType.BODY)) {
                            bodyParamType = endpoint.getMethodInfo().parameterType(i);
                            break;
                        }
                    }

                    if (bodyParamType instanceof ArrayType) {
                        org.jboss.jandex.Type component = ((ArrayType) bodyParamType).component();
                        if (component instanceof PrimitiveType) {
                            base64 = component.asPrimitiveType().equals(PrimitiveType.BYTE);
                        }
                    }

                    recorder.registerHttpIngress(moduleBuilder.getModuleName(), verbName, base64);

                    StringBuilder pathBuilder = new StringBuilder();
                    if (endpoint.getBasicResourceClassInfo().getPath() != null) {
                        pathBuilder.append(endpoint.getBasicResourceClassInfo().getPath());
                    }
                    if (endpoint.getResourceMethod().getPath() != null && !endpoint.getResourceMethod().getPath().isEmpty()) {
                        boolean builderEndsSlash = pathBuilder.charAt(pathBuilder.length() - 1) == '/';
                        boolean pathStartsSlash = endpoint.getResourceMethod().getPath().startsWith("/");
                        if (builderEndsSlash && pathStartsSlash) {
                            pathBuilder.setLength(pathBuilder.length() - 1);
                        } else if (!builderEndsSlash && !pathStartsSlash) {
                            pathBuilder.append('/');
                        }
                        pathBuilder.append(endpoint.getResourceMethod().getPath());
                    }
                    String path = pathBuilder.toString();
                    URITemplate template = new URITemplate(path, false);
                    List<IngressPathComponent> pathComponents = new ArrayList<>();
                    for (var i : template.components) {
                        if (i.type == URITemplate.Type.CUSTOM_REGEX) {
                            throw new RuntimeException(
                                    "Invalid path " + path + " on HTTP endpoint: " + endpoint.getActualClassInfo().name() + "."
                                            + ModuleBuilder.methodToName(endpoint.getMethodInfo())
                                            + " FTL does not support custom regular expressions");
                        } else if (i.type == URITemplate.Type.LITERAL) {
                            for (var part : i.literalText.split("/")) {
                                if (part.isEmpty()) {
                                    continue;
                                }
                                pathComponents.add(IngressPathComponent.newBuilder()
                                        .setIngressPathLiteral(IngressPathLiteral.newBuilder().setText(part))
                                        .build());
                            }
                        } else {
                            pathComponents.add(IngressPathComponent.newBuilder()
                                    .setIngressPathParameter(IngressPathParameter.newBuilder().setName(i.name))
                                    .build());
                        }
                    }

                    //TODO: process path properly
                    MetadataIngress.Builder ingressBuilder = MetadataIngress.newBuilder()
                            .setType("http")
                            .setMethod(endpoint.getResourceMethod().getHttpMethod());
                    for (var i : pathComponents) {
                        ingressBuilder.addPath(i);
                    }
                    Metadata ingressMetadata = Metadata.newBuilder()
                            .setIngress(ingressBuilder
                                    .build())
                            .build();
                    Type requestTypeParam = moduleBuilder.buildType(bodyParamType, true, Nullability.NOT_NULL);
                    Type responseTypeParam = moduleBuilder.buildType(endpoint.getMethodInfo().returnType(), true,
                            Nullability.NOT_NULL);
                    Type stringType = Type.newBuilder().setString(xyz.block.ftl.v1.schema.String.newBuilder().build()).build();
                    Type pathParamType = Type.newBuilder()
                            .setMap(xyz.block.ftl.v1.schema.Map.newBuilder().setKey(stringType)
                                    .setValue(stringType))
                            .build();
                    moduleBuilder
                            .addDecls(Decl.newBuilder().setVerb(xyz.block.ftl.v1.schema.Verb.newBuilder()
                                    .addMetadata(ingressMetadata)
                                    .setName(verbName)
                                    .setExport(true)
                                    .setRequest(Type.newBuilder()
                                            .setRef(Ref.newBuilder().setModule(ModuleBuilder.BUILTIN)
                                                    .setName(HttpRequest.class.getSimpleName())
                                                    .addTypeParameters(requestTypeParam)
                                                    .addTypeParameters(pathParamType)
                                                    .addTypeParameters(Type.newBuilder()
                                                            .setMap(xyz.block.ftl.v1.schema.Map.newBuilder().setKey(stringType)
                                                                    .setValue(Type.newBuilder()
                                                                            .setArray(
                                                                                    Array.newBuilder().setElement(stringType)))
                                                                    .build())))
                                            .build())
                                    .setResponse(Type.newBuilder()
                                            .setRef(Ref.newBuilder().setModule(ModuleBuilder.BUILTIN)
                                                    .setName(HttpResponse.class.getSimpleName())
                                                    .addTypeParameters(responseTypeParam)
                                                    .addTypeParameters(Type.newBuilder().setUnit(Unit.newBuilder())))
                                            .build()))
                                    .build());
                }
            }
        });

    }
}
