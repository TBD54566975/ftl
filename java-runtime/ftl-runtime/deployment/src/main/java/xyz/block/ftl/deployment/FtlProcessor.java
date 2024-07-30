package xyz.block.ftl.deployment;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.arc.runtime.AdditionalBean;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.FeatureBuildItem;
import io.quarkus.deployment.pkg.builditem.OutputTargetBuildItem;
import io.quarkus.deployment.recording.RecorderContext;
import io.quarkus.grpc.deployment.BindableServiceBuildItem;
import org.checkerframework.checker.units.qual.A;
import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.DotName;
import org.jboss.jandex.IndexView;
import org.jboss.jandex.VoidType;
import xyz.block.ftl.Verb;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.VerbHandler;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.v1.schema.*;
import xyz.block.ftl.v1.schema.Float;
import xyz.block.ftl.v1.schema.Module;

import java.io.IOException;
import java.lang.String;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;

class FtlProcessor {

    private static final String SCHEMA_OUT = "schema.pb";
    private static final String FEATURE = "ftl-java-runtime";
    public static final String HACK_MODULE_NAME = "demomodule";

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
    public  void registerVerbs(CombinedIndexBuildItem index,
                               FTLRecorder recorder,
                               BuildProducer<AdditionalBeanBuildItem> additionalBeanBuildItemBuildProducer,
                               OutputTargetBuildItem outputTargetBuildItem,
                               RecorderContext recorderContext) throws IOException {
        Module.Builder moduleBuilder = Module.newBuilder()
            .setName("testmodule")
            .setBuiltin(false);
        Map<String, Ref> dataElements = new HashMap<>();
        var beans = AdditionalBeanBuildItem.builder().setUnremovable();
        for (var verb : index.getIndex().getAnnotations(DotName.createSimple(Verb.class))) {
            AnnotationValue exportedValue = verb.value("exported");
            boolean exported = exportedValue != null && exportedValue.asBoolean();
            var method = verb.target().asMethod();
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);
            org.jboss.jandex.Type methodParamType;
            if (method.parametersCount() == 0) {
                methodParamType = VoidType.VOID;
            } else if (method.parametersCount() == 1) {
                methodParamType = method.parameters().get(0).type();
            } else  {
                throw new RuntimeException("@Verb methods must only have a single parameter: " + method.declaringClass().name() + "." + method.name());
            }
            recorder.registerVerb(HACK_MODULE_NAME, method.name(), recorderContext.classProxy(methodParamType.toString()), method.name(), recorderContext.classProxy(className));
            moduleBuilder
                    .addDecls(Decl.newBuilder().setVerb(xyz.block.ftl.v1.schema.Verb.newBuilder()
                            .setName(method.name())
                            .setExport(exported)
                            .setRequest(buildType(index.getComputingIndex(), methodParamType, dataElements, moduleBuilder))
                            .setRequest(buildType(index.getComputingIndex(), method.returnType(), dataElements, moduleBuilder)))
                            .build());
        }
        additionalBeanBuildItemBuildProducer.produce(beans.build());
        Path output = outputTargetBuildItem.getOutputDirectory().resolve(SCHEMA_OUT);
        try (var out = Files.newOutputStream(output)) {
            moduleBuilder.build().writeTo(out);
        }

    }

    private Type buildType(IndexView index, org.jboss.jandex.Type type, Map<String, Ref> dataElements, Module.Builder moduleBuilder) {
        switch (type.kind()) {
            case PRIMITIVE -> {
                var prim = type.asPrimitiveType();
                switch (prim.primitive()) {
                    case INT, LONG , BYTE, SHORT-> {
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
            }

        }

        throw new RuntimeException("NOT YET IMPLEMENTED");
    }
}
