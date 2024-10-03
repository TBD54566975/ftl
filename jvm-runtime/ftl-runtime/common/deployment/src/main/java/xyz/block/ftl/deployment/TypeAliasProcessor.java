package xyz.block.ftl.deployment;

import java.util.HashMap;
import java.util.Map;

import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.Type;
import org.jboss.jandex.TypeVariable;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;

public class TypeAliasProcessor {

    @BuildStep
    public void processTypeAlias(CombinedIndexBuildItem index,
            BuildProducer<SchemaContributorBuildItem> schemaContributorBuildItemBuildProducer,
            BuildProducer<AdditionalBeanBuildItem> additionalBeanBuildItem,
            BuildProducer<TypeAliasBuildItem> typeAliasBuildItemBuildProducer) {
        var beans = new AdditionalBeanBuildItem.Builder().setUnremovable();
        for (var mapper : index.getIndex().getAnnotations(FTLDotNames.TYPE_ALIAS)) {
            boolean exported = mapper.target().hasAnnotation(FTLDotNames.EXPORT);
            // This may or may not be the actual mapper, it may be a subclass

            var mapperClass = mapper.target().asClass();
            var actualMapper = mapperClass;

            Type t = null;
            Type s = null;
            if (mapperClass.isInterface()) {
                for (var i : mapperClass.interfaceTypes()) {
                    if (i.name().equals(FTLDotNames.TYPE_ALIAS_MAPPER)) {
                        t = i.asParameterizedType().arguments().get(0);
                        s = i.asParameterizedType().arguments().get(1);
                        break;
                    }
                }
                var implementations = index.getComputingIndex().getAllKnownImplementors(mapperClass.name());
                if (implementations.isEmpty()) {
                    continue;
                }
                if (implementations.size() > 1) {
                    throw new RuntimeException(
                            "Multiple implementations of " + mapperClass.name() + " found: " + implementations);
                }
                actualMapper = implementations.iterator().next();
            }

            //TODO: this is a bit hacky and won't work for complex heirachies
            // it is enough to get us going through
            for (var i : actualMapper.interfaceTypes()) {
                if (i.name().equals(FTLDotNames.TYPE_ALIAS_MAPPER)) {
                    t = i.asParameterizedType().arguments().get(0);
                    s = i.asParameterizedType().arguments().get(1);
                    break;
                } else if (i.name().equals(mapperClass.name())) {
                    if (t instanceof TypeVariable) {
                        t = i.asParameterizedType().arguments().get(0);
                    }
                    if (s instanceof TypeVariable) {
                        s = i.asParameterizedType().arguments().get(1);
                    }
                    break;
                }
            }

            beans.addBeanClass(actualMapper.name().toString());
            var finalT = t;
            var finalS = s;
            String module = mapper.value("module") == null ? "" : mapper.value("module").asString();
            String name = mapper.value("name").asString();
            typeAliasBuildItemBuildProducer.produce(new TypeAliasBuildItem(name, module, t, s, exported));
            if (module.isEmpty()) {
                Map<String, String> languageMappings = new HashMap<>();
                AnnotationValue languageTypeMappingsValue = mapper.value("languageTypeMappings");
                if (languageTypeMappingsValue != null) {
                    for (var lang : languageTypeMappingsValue.asArrayList()) {
                        languageMappings.put(lang.asNested().value("language").asString(),
                                lang.asNested().value("type").asString());
                    }
                }
                schemaContributorBuildItemBuildProducer.produce(new SchemaContributorBuildItem(moduleBuilder -> moduleBuilder
                        .registerTypeAlias(name, finalT, finalS, exported, languageMappings)));
            }

        }
        additionalBeanBuildItem.produce(beans.build());

    }

}
