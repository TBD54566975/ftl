package xyz.block.ftl.deployment;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.arc.runtime.AdditionalBean;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.FeatureBuildItem;
import io.quarkus.grpc.deployment.BindableServiceBuildItem;
import org.checkerframework.checker.units.qual.A;
import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.DotName;
import xyz.block.ftl.Verb;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.VerbHandler;
import xyz.block.ftl.runtime.VerbRegistry;

class FtlProcessor {

    private static final String FEATURE = "ftl-java-runtime";

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
                               BuildProducer<AdditionalBeanBuildItem> additionalBeanBuildItemBuildProducer) {
        var beans = AdditionalBeanBuildItem.builder();
        for (var verb : index.getIndex().getAnnotations(DotName.createSimple(Verb.class))) {
            AnnotationValue exportedValue = verb.value("exported");
            boolean exported = exportedValue != null && exportedValue.asBoolean();
            var method = verb.target().asMethod();
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);
            if (method.parametersCount() != 1) {
                throw new RuntimeException("@Verb methods must only have a single parameter: " + method.declaringClass().name() + "." + method.name());
            }
            recorder.registerVerb("demomodule", method.name(), method.parameters().get(0).type().asClassType().name().toString(), method.returnType().asClassType().name().toString(), method.name(), className);
        }
        additionalBeanBuildItemBuildProducer.produce(beans.build());
    }
}
