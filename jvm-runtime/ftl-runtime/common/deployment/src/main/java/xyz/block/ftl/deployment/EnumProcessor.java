package xyz.block.ftl.deployment;

import java.lang.reflect.Field;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.DotName;
import org.jboss.jandex.PrimitiveType;
import org.jboss.jandex.Type;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import xyz.block.ftl.v1.schema.Decl;
import xyz.block.ftl.v1.schema.Enum;
import xyz.block.ftl.v1.schema.EnumVariant;
import xyz.block.ftl.v1.schema.Int;
import xyz.block.ftl.v1.schema.IntValue;
import xyz.block.ftl.v1.schema.StringValue;
import xyz.block.ftl.v1.schema.Value;

public class EnumProcessor {

    private static final Logger log = LoggerFactory.getLogger(EnumProcessor.class);

    @BuildStep
    SchemaContributorBuildItem handleEnums(CombinedIndexBuildItem index) {
        var enumAnnotations = index.getIndex().getAnnotations(FTLDotNames.ENUM);
        log.info("Processing {} enum annotations into build items", enumAnnotations.size());
        List<Decl> decls = new ArrayList<>();
        try {
            for (var enumAnnotation : enumAnnotations) {
                boolean exported = enumAnnotation.target().hasAnnotation(FTLDotNames.EXPORT);
                ClassInfo enumClassInfo = enumAnnotation.target().asClass();
                Enum.Builder enumBuilder = Enum.newBuilder()
                        .setName(enumClassInfo.simpleName())
                        .setExport(exported);
                if (enumClassInfo.isEnum()) {
                    // Value enums must have a type
                    Type type = enumClassInfo.field("value").type();
                    xyz.block.ftl.v1.schema.Type.Builder typeBuilder = xyz.block.ftl.v1.schema.Type.newBuilder();
                    if (type == PrimitiveType.LONG) {
                        typeBuilder.setInt(Int.newBuilder().build()).build();
                    } else {
                        typeBuilder.setString(xyz.block.ftl.v1.schema.String.newBuilder().build());
                    }
                    enumBuilder.setType(typeBuilder.build());

                    Class<?> enumClass = Class.forName(enumClassInfo.name().toString(), false,
                            Thread.currentThread().getContextClassLoader());
                    for (var constant : enumClass.getEnumConstants()) {
                        Field value = constant.getClass().getDeclaredField("value");
                        value.setAccessible(true);
                        Value.Builder valueBuilder = Value.newBuilder();
                        if (type == PrimitiveType.LONG) {
                            long aLong = value.getLong(constant);
                            valueBuilder.setIntValue(IntValue.newBuilder().setValue(aLong).build());
                        } else if (type.name().equals(DotName.STRING_NAME)) {
                            String aString = (String) value.get(constant);
                            valueBuilder.setStringValue(StringValue.newBuilder().setValue(aString).build());
                        } else {
                            throw new RuntimeException("Unsupported enum value type: " + value.getType());
                        }
                        EnumVariant variant = EnumVariant.newBuilder()
                                .setName(constant.toString())
                                .setValue(valueBuilder)
                                .build();
                        enumBuilder.addVariants(variant);
                    }
                    // TODO move outside if
                    decls.add(Decl.newBuilder().setEnum(enumBuilder).build());
                } else {
                    // Type enums
                    // TODO
                }

            }
            return new SchemaContributorBuildItem(decls);
        } catch (ClassNotFoundException | NoSuchFieldException | IllegalAccessException e) {
            throw new RuntimeException(e);
        }
    }

}
