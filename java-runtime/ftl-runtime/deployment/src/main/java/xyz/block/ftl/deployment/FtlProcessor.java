package xyz.block.ftl.deployment;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.arc.processor.DotNames;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.ApplicationInfoBuildItem;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.FeatureBuildItem;
import io.quarkus.deployment.pkg.builditem.OutputTargetBuildItem;
import io.quarkus.deployment.recording.RecorderContext;
import io.quarkus.grpc.deployment.BindableServiceBuildItem;
import io.quarkus.resteasy.reactive.server.deployment.ResteasyReactiveResourceMethodEntriesBuildItem;
import org.jboss.jandex.DotName;
import org.jboss.jandex.IndexView;
import org.jboss.jandex.VoidType;
import org.jboss.resteasy.reactive.common.model.MethodParameter;
import org.jboss.resteasy.reactive.common.model.ParameterType;
import org.jboss.resteasy.reactive.server.mapping.URITemplate;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;
import xyz.block.ftl.runtime.FTLHttpHandler;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.VerbHandler;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.builtin.HttpRequest;
import xyz.block.ftl.runtime.builtin.HttpResponse;
import xyz.block.ftl.v1.schema.Array;
import xyz.block.ftl.v1.schema.Bool;
import xyz.block.ftl.v1.schema.Data;
import xyz.block.ftl.v1.schema.Decl;
import xyz.block.ftl.v1.schema.Field;
import xyz.block.ftl.v1.schema.Float;
import xyz.block.ftl.v1.schema.IngressPathComponent;
import xyz.block.ftl.v1.schema.IngressPathLiteral;
import xyz.block.ftl.v1.schema.IngressPathParameter;
import xyz.block.ftl.v1.schema.Int;
import xyz.block.ftl.v1.schema.Metadata;
import xyz.block.ftl.v1.schema.MetadataIngress;
import xyz.block.ftl.v1.schema.Module;
import xyz.block.ftl.v1.schema.Optional;
import xyz.block.ftl.v1.schema.Ref;
import xyz.block.ftl.v1.schema.Type;
import xyz.block.ftl.v1.schema.Unit;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.PosixFilePermission;
import java.util.ArrayList;
import java.util.EnumSet;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

class FtlProcessor {

    private static final String SCHEMA_OUT = "schema.pb";
    private static final String FEATURE = "ftl-java-runtime";
    public static final DotName EXPORT = DotName.createSimple(Export.class);
    public static final DotName VERB = DotName.createSimple(Verb.class);
    public static final String BUILTIN = "builtin";

    @BuildStep
    FeatureBuildItem feature() {
        return new FeatureBuildItem(FEATURE);
    }

    @BuildStep
    BindableServiceBuildItem verbService() {
        var ret = new BindableServiceBuildItem(DotName.createSimple(VerbHandler.class));
        ret.registerBlockingMethod("call");
        ret.registerBlockingMethod("publishEvent");
        ret.registerBlockingMethod("sendFSMEvent");
        ret.registerBlockingMethod("acquireLease");
        ret.registerBlockingMethod("getModuleContext");
        ret.registerBlockingMethod("ping");
        return ret;
    }

    @BuildStep
    AdditionalBeanBuildItem beans() {
        return AdditionalBeanBuildItem.builder()
                .addBeanClasses(VerbHandler.class, VerbRegistry.class, FTLHttpHandler.class)
                .setUnremovable().build();
    }

    @BuildStep
    AdditionalBeanBuildItem verbBeans(CombinedIndexBuildItem index) {

        var beans = AdditionalBeanBuildItem.builder().setUnremovable();
        for (var verb : index.getIndex().getAnnotations(VERB)) {
            beans.addBeanClasses(verb.target().asMethod().declaringClass().name().toString());
        }
        return beans.build();
    }

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    public void registerVerbs(CombinedIndexBuildItem index,
                              FTLRecorder recorder,
                              ApplicationInfoBuildItem applicationInfoBuildItem,
                              OutputTargetBuildItem outputTargetBuildItem,
                              RecorderContext recorderContext,
                              ResteasyReactiveResourceMethodEntriesBuildItem restEndpoints) throws IOException {
        Module.Builder moduleBuilder = Module.newBuilder()
                .setName(applicationInfoBuildItem.getName())
                .setBuiltin(false);
        Map<TypeKey, Ref> dataElements = new HashMap<>();
        var beans = AdditionalBeanBuildItem.builder().setUnremovable();
        for (var verb : index.getIndex().getAnnotations(VERB)) {
            boolean exported = verb.target().hasAnnotation(EXPORT);
            var method = verb.target().asMethod();
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);
            org.jboss.jandex.Type methodParamType;
            if (method.parametersCount() == 0) {
                methodParamType = VoidType.VOID;
            } else if (method.parametersCount() == 1) {
                methodParamType = method.parameters().get(0).type();
            } else {
                throw new RuntimeException("@Verb methods must only have a single parameter: " + method.declaringClass().name() + "." + method.name());
            }
            recorder.registerVerb(applicationInfoBuildItem.getName(), method.name(), recorderContext.classProxy(methodParamType.toString()), method.name(), recorderContext.classProxy(className));
            moduleBuilder
                    .addDecls(Decl.newBuilder().setVerb(xyz.block.ftl.v1.schema.Verb.newBuilder()
                                    .setName(method.name())
                                    .setExport(exported)
                                    .setRequest(buildType(index.getComputingIndex(), methodParamType, dataElements, moduleBuilder))
                                    .setResponse(buildType(index.getComputingIndex(), method.returnType(), dataElements, moduleBuilder)))
                            .build());
        }

        //TODO: make this composable so it is not just one big method, build items should contribute to the schema
        for (var endpoint : restEndpoints.getEntries()) {
            //TODO: naming
            var verbName = endpoint.getMethodInfo().name();
            recorder.registerHttpIngress(applicationInfoBuildItem.getName(), verbName);

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

            StringBuilder pathBuilder = new StringBuilder();
            if (endpoint.getBasicResourceClassInfo().getPath() != null) {
                pathBuilder.append(endpoint.getBasicResourceClassInfo().getPath());
            }
            if (endpoint.getResourceMethod().getPath() != null && !endpoint.getResourceMethod().getPath().isEmpty()) {
                if (pathBuilder.charAt(pathBuilder.length() - 1) != '/' && !endpoint.getResourceMethod().getPath().startsWith("/")) {
                    pathBuilder.append('/');
                }
                pathBuilder.append(endpoint.getResourceMethod().getPath());
            }
            String path = pathBuilder.toString();
            URITemplate template = new URITemplate(path, false);
            List<IngressPathComponent> pathComponents = new ArrayList<>();
            for (var i : template.components) {
                if (i.type == URITemplate.Type.CUSTOM_REGEX) {
                    throw new RuntimeException("Invalid path " + path + " on HTTP endpoint: " + endpoint.getActualClassInfo().name() + "." + endpoint.getMethodInfo().name() + " FTL does not support custom regular expressions");
                } else if (i.type == URITemplate.Type.LITERAL) {
                    pathComponents.add(IngressPathComponent.newBuilder().setIngressPathLiteral(IngressPathLiteral.newBuilder().setText(i.literalText)).build());
                } else {
                    pathComponents.add(IngressPathComponent.newBuilder().setIngressPathParameter(IngressPathParameter.newBuilder().setName(i.name)).build());
                }
            }

            //TODO: process path properly
            MetadataIngress.Builder ingressBuilder = MetadataIngress.newBuilder()
                    .setMethod(endpoint.getResourceMethod().getHttpMethod());
            for (var i : pathComponents) {
                ingressBuilder.addPath(i);
            }
            Metadata ingressMetadata = Metadata.newBuilder()
                    .setIngress(ingressBuilder
                            .build())
                    .build();
            Type requestTypeParam = buildType(index.getComputingIndex(), bodyParamType, dataElements, moduleBuilder);
            Type responseTypeParam = buildType(index.getComputingIndex(), endpoint.getMethodInfo().returnType(), dataElements, moduleBuilder);
            moduleBuilder
                    .addDecls(Decl.newBuilder().setVerb(xyz.block.ftl.v1.schema.Verb.newBuilder()
                                    .addMetadata(ingressMetadata)
                                    .setName(verbName)
                                    .setExport(true)
                                    .setRequest(Type.newBuilder().setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpRequest.class.getSimpleName()).addTypeParameters(requestTypeParam)).build())
                                    .setResponse(Type.newBuilder().setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpResponse.class.getSimpleName()).addTypeParameters(responseTypeParam).addTypeParameters(Type.newBuilder().setUnit(Unit.newBuilder()))).build()))
                            .build());
        }

        Path output = outputTargetBuildItem.getOutputDirectory().resolve(SCHEMA_OUT);
        try (var out = Files.newOutputStream(output)) {
            moduleBuilder.build().writeTo(out);
        }

        output = outputTargetBuildItem.getOutputDirectory().resolve("main");
        try (var out = Files.newOutputStream(output)) {
            out.write("""
                    #!/bin/bash
                    exec java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005 -jar quarkus-app/quarkus-run.jar""".getBytes(StandardCharsets.UTF_8));
        }
        var perms = Files.getPosixFilePermissions(output);
        EnumSet<PosixFilePermission> newPerms = EnumSet.copyOf(perms);
        newPerms.add(PosixFilePermission.GROUP_EXECUTE);
        newPerms.add(PosixFilePermission.OWNER_EXECUTE);
        Files.setPosixFilePermissions(output, newPerms);
    }

    private Type buildType(IndexView index, org.jboss.jandex.Type type, Map<TypeKey, Ref> dataElements, Module.Builder moduleBuilder) {
        switch (type.kind()) {
            case PRIMITIVE -> {
                var prim = type.asPrimitiveType();
                switch (prim.primitive()) {
                    case INT, LONG, BYTE, SHORT -> {
                        return Type.newBuilder().setInt(Int.newBuilder().build()).build();
                    }
                    case FLOAT, DOUBLE -> {
                        return Type.newBuilder().setFloat(Float.newBuilder().build()).build();
                    }
                    case BOOLEAN -> {
                        return Type.newBuilder().setBool(Bool.newBuilder().build()).build();
                    }
                    case CHAR -> {
                        return Type.newBuilder().setString(xyz.block.ftl.v1.schema.String.newBuilder().build()).build();
                    }
                    default -> throw new RuntimeException("unknown primitive type: " + prim.primitive());
                }
            }
            case VOID -> {
                return Type.newBuilder().setUnit(Unit.newBuilder().build()).build();
            }
            case ARRAY -> {
                return Type.newBuilder().setArray(Array.newBuilder().setElement(buildType(index, type.asArrayType().componentType(), dataElements, moduleBuilder)).build()).build();
            }
            case CLASS -> {
                var clazz = type.asClassType();
                if (clazz.name().equals(DotName.STRING_NAME)) {
                    return Type.newBuilder().setString(xyz.block.ftl.v1.schema.String.newBuilder().build()).build();
                }
                var existing = dataElements.get(new TypeKey(clazz.name().toString()));
                if (existing != null) {
                    return Type.newBuilder().setRef(existing).build();
                }
                Data.Builder data = Data.newBuilder();
                data.setName(clazz.name().local());
                data.setExport(type.hasAnnotation(EXPORT));
                buildDataElement(data, index, clazz.name(), dataElements, moduleBuilder);
                moduleBuilder.addDecls(Decl.newBuilder().setData(data).build());
                Ref ref = Ref.newBuilder().setName(data.getName()).setModule(moduleBuilder.getName()).build();
                dataElements.put(new TypeKey(clazz.name().toString()), ref);
                return Type.newBuilder().setRef(ref).build();
            }
            case PARAMETERIZED_TYPE -> {
                var paramType = type.asParameterizedType();
                if (paramType.name().equals(DotName.createSimple(List.class))) {
                    return Type.newBuilder().setArray(Array.newBuilder().setElement(buildType(index, paramType.arguments().get(0), dataElements, moduleBuilder))).build();
                } else if (paramType.name().equals(DotName.createSimple(Map.class))) {
                    return Type.newBuilder().setMap(xyz.block.ftl.v1.schema.Map.newBuilder()
                                    .setKey(buildType(index, paramType.arguments().get(0), dataElements, moduleBuilder))
                                    .setValue(buildType(index, paramType.arguments().get(0), dataElements, moduleBuilder)))
                            .build();
                } else if (paramType.name().equals(DotNames.OPTIONAL)) {
                    return Type.newBuilder().setOptional(Optional.newBuilder().setType(buildType(index, paramType.arguments().get(0), dataElements, moduleBuilder))).build();
                } else if (paramType.name().equals(DotName.createSimple(HttpRequest.class))) {
                    return Type.newBuilder().setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpRequest.class.getSimpleName()).addTypeParameters(buildType(index, paramType.arguments().get(0), dataElements, moduleBuilder))).build();
                } else if (paramType.name().equals(DotName.createSimple(HttpResponse.class))) {
                    return Type.newBuilder().setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpResponse.class.getSimpleName()).addTypeParameters(buildType(index, paramType.arguments().get(0), dataElements, moduleBuilder)).addTypeParameters(Type.newBuilder().setUnit(Unit.newBuilder().build()))).build();
                }
            }
        }

        throw new RuntimeException("NOT YET IMPLEMENTED");
    }

    private void buildDataElement(Data.Builder data, IndexView index, DotName className, Map<TypeKey, Ref> dataElements, Module.Builder moduleBuilder) {
        if (className == null || className.equals(DotName.OBJECT_NAME)) {
            return;
        }
        var clazz = index.getClassByName(className);
        if (clazz == null) {
            return;
        }
        //TODO: handle getters and setters properly, also Jackson annotations etc
        for (var field : clazz.fields()) {
            data.addFields(Field.newBuilder().setName(field.name()).setType(buildType(index, field.type(), dataElements, moduleBuilder)).build());
        }
        buildDataElement(data, index, clazz.superName(), dataElements, moduleBuilder);
    }

    private record TypeKey(String name, String... typeParams) {

    }
}
