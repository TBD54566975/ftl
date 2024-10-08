package xyz.block.ftl.deployment;

import static org.jboss.jandex.PrimitiveType.Primitive.BYTE;
import static org.jboss.jandex.PrimitiveType.Primitive.INT;
import static org.jboss.jandex.PrimitiveType.Primitive.LONG;
import static org.jboss.jandex.PrimitiveType.Primitive.SHORT;
import static xyz.block.ftl.deployment.FTLDotNames.ENUM_HOLDER;
import static xyz.block.ftl.deployment.FTLDotNames.GENERATED_REF;

import java.lang.reflect.Field;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import java.util.function.Consumer;

import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.ClassType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.FieldInfo;
import org.jboss.jandex.Type;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.v1.schema.Decl;
import xyz.block.ftl.v1.schema.Enum;
import xyz.block.ftl.v1.schema.EnumVariant;
import xyz.block.ftl.v1.schema.Int;
import xyz.block.ftl.v1.schema.IntValue;
import xyz.block.ftl.v1.schema.StringValue;
import xyz.block.ftl.v1.schema.TypeValue;
import xyz.block.ftl.v1.schema.Value;

public class EnumProcessor {

    private static final Logger log = LoggerFactory.getLogger(EnumProcessor.class);

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    SchemaContributorBuildItem handleEnums(CombinedIndexBuildItem index, FTLRecorder recorder) {
        var enumAnnotations = index.getIndex().getAnnotations(FTLDotNames.ENUM);
        log.info("Processing {} enum annotations into decls", enumAnnotations.size());

        return new SchemaContributorBuildItem(new Consumer<ModuleBuilder>() {
            @Override
            public void accept(ModuleBuilder moduleBuilder) {
                List<Decl> decls = new ArrayList<>();
                try {
                    for (var enumAnnotation : enumAnnotations) {
                        boolean exported = enumAnnotation.target().hasAnnotation(FTLDotNames.EXPORT);
                        ClassInfo classInfo = enumAnnotation.target().asClass();
                        Class<?> clazz = Class.forName(classInfo.name().toString(), false,
                                Thread.currentThread().getContextClassLoader());
                        var isLocalToModule = !classInfo.hasDeclaredAnnotation(GENERATED_REF);
                        Enum.Builder enumBuilder = Enum.newBuilder()
                                .setName(classInfo.simpleName())
                                .setExport(exported);
                        if (classInfo.isEnum()) {
                            recorder.registerEnum(clazz);
                            if (isLocalToModule) {
                                decls.add(extractValueEnum(classInfo, clazz, enumBuilder));
                            }
                        } else {
                            // Type enums
                            var variants = index.getComputingIndex().getAllKnownImplementors(classInfo.name());
                            var variantClasses = new ArrayList<Class<?>>();
                            if (variants.isEmpty()) {
                                throw new RuntimeException("No variants found for enum: " + enumBuilder.getName());
                            }
                            for (var variant : variants) {
                                var isVariantLocalToModule = !variant.hasDeclaredAnnotation(GENERATED_REF);
                                Type variantType;
                                if (variant.hasAnnotation(ENUM_HOLDER)) {
                                    // Enum value holder class
                                    FieldInfo valueField = variant.field("value");
                                    if (valueField == null) {
                                        throw new RuntimeException("Enum variant must have a 'value' field: " + variant.name());
                                    }
                                    variantType = valueField.type();
                                } else {
                                    // Class is the enum variant type
                                    variantType = ClassType.builder(variant.name()).build();
                                    Class<?> variantClazz = Class.forName(variantType.name().toString(), false,
                                            Thread.currentThread().getContextClassLoader());
                                    variantClasses.add(variantClazz);
                                }
                                if (isVariantLocalToModule) {
                                    xyz.block.ftl.v1.schema.Type declType = moduleBuilder.buildType(variantType, exported,
                                            Nullability.NOT_NULL);
                                    TypeValue typeValue = TypeValue.newBuilder().setValue(declType).build();

                                    EnumVariant.Builder variantBuilder = EnumVariant.newBuilder()
                                            .setName(variant.simpleName())
                                            .setValue(Value.newBuilder().setTypeValue(typeValue).build());
                                    enumBuilder.addVariants(variantBuilder.build());
                                }
                            }
                            if (isLocalToModule) {
                                decls.add(Decl.newBuilder().setEnum(enumBuilder).build());
                            }
                            recorder.registerEnum(clazz, variantClasses);
                        }
                    }
                    for (var decl : decls) {
                        moduleBuilder.addDecls(decl);
                    }
                } catch (ClassNotFoundException | NoSuchFieldException | IllegalAccessException e) {
                    throw new RuntimeException(e);
                }
            }
        });
    }

    private Decl extractValueEnum(ClassInfo classInfo, Class<?> clazz, Enum.Builder enumBuilder)
            throws ClassNotFoundException, NoSuchFieldException, IllegalAccessException {
        // Value enums must have a type
        FieldInfo valueField = classInfo.field("value");
        if (valueField == null) {
            throw new RuntimeException("Enum must have a 'value' field: " + classInfo.name());
        }
        Type type = valueField.type();
        xyz.block.ftl.v1.schema.Type.Builder typeBuilder = xyz.block.ftl.v1.schema.Type.newBuilder();
        if (isInt(type)) {
            typeBuilder.setInt(Int.newBuilder().build()).build();
        } else if (type.name().equals(DotName.STRING_NAME)) {
            typeBuilder.setString(xyz.block.ftl.v1.schema.String.newBuilder().build());
        } else {
            throw new RuntimeException(
                    "Enum value type must be String, int, long, short, or byte: " + classInfo.name());
        }
        enumBuilder.setType(typeBuilder.build());

        for (var constant : clazz.getEnumConstants()) {
            Field value = constant.getClass().getDeclaredField("value");
            value.setAccessible(true);
            Value.Builder valueBuilder = Value.newBuilder();
            if (isInt(type)) {
                long aLong = value.getLong(constant);
                valueBuilder.setIntValue(IntValue.newBuilder().setValue(aLong).build());
            } else {
                String aString = (String) value.get(constant);
                valueBuilder.setStringValue(StringValue.newBuilder().setValue(aString).build());
            }
            EnumVariant variant = EnumVariant.newBuilder()
                    .setName(constant.toString())
                    .setValue(valueBuilder)
                    .build();
            enumBuilder.addVariants(variant);
        }
        return Decl.newBuilder().setEnum(enumBuilder).build();
    }

    private boolean isInt(Type type) {
        if (type.kind() != Type.Kind.PRIMITIVE) {
            return false;
        }
        return Set.of(INT, LONG, BYTE, SHORT).contains(type.asPrimitiveType().primitive());
    }

}
