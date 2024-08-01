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
import org.jboss.jandex.*;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.VerbHandler;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.v1.schema.Float;
import xyz.block.ftl.v1.schema.Module;
import xyz.block.ftl.v1.schema.*;
import xyz.block.ftl.v1.schema.Type;

import java.io.File;
import java.io.IOException;
import java.lang.String;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.PosixFilePermission;
import java.util.EnumSet;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

class FtlProcessor {

    private static final String SCHEMA_OUT = "schema.pb";
    private static final String FEATURE = "ftl-java-runtime";
    public static final DotName EXPORT = DotName.createSimple(Export.class);
    public static final DotName VERB = DotName.createSimple(Verb.class);

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
                .addBeanClasses(VerbHandler.class, VerbRegistry.class)
                .setUnremovable().build();
    }

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    public void registerVerbs(CombinedIndexBuildItem index,
                              FTLRecorder recorder,
                              ApplicationInfoBuildItem applicationInfoBuildItem,
                              BuildProducer<AdditionalBeanBuildItem> additionalBeanBuildItemBuildProducer,
                              OutputTargetBuildItem outputTargetBuildItem,
                              RecorderContext recorderContext) throws IOException {
        Module.Builder moduleBuilder = Module.newBuilder()
                .setName(applicationInfoBuildItem.getName())
                .setBuiltin(false);
        Map<String, Ref> dataElements = new HashMap<>();
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
        additionalBeanBuildItemBuildProducer.produce(beans.build());
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

    private Type buildType(IndexView index, org.jboss.jandex.Type type, Map<String, Ref> dataElements, Module.Builder moduleBuilder) {
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
                var existing = dataElements.get(clazz.name().toString());
                if (existing != null) {
                    return Type.newBuilder().setRef(existing).build();
                }
                Data.Builder data = Data.newBuilder();
                data.setName(clazz.name().local());
                data.setExport(type.hasAnnotation(EXPORT));
                buildDataElement(data, index, clazz.name(), dataElements, moduleBuilder);
                moduleBuilder.addDecls(Decl.newBuilder().setData(data).build());
                Ref ref = Ref.newBuilder().setName(data.getName()).setModule(moduleBuilder.getName()).build();
                dataElements.put(clazz.name().toString(), ref);
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
                }
            }
        }

        throw new RuntimeException("NOT YET IMPLEMENTED");
    }

    private void buildDataElement(Data.Builder data, IndexView index, DotName className, Map<String, Ref> dataElements, Module.Builder moduleBuilder) {
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
}
