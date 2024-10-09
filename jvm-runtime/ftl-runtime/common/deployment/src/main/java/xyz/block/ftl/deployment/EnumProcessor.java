package xyz.block.ftl.deployment;

import static org.jboss.jandex.PrimitiveType.Primitive.BYTE;
import static org.jboss.jandex.PrimitiveType.Primitive.INT;
import static org.jboss.jandex.PrimitiveType.Primitive.LONG;
import static org.jboss.jandex.PrimitiveType.Primitive.SHORT;
import static xyz.block.ftl.deployment.FTLDotNames.ENUM_HOLDER;
import static xyz.block.ftl.deployment.FTLDotNames.GENERATED_REF;

import java.lang.reflect.Field;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Set;
import java.util.function.Consumer;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.ClassType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.FieldInfo;
import org.jboss.jandex.PrimitiveType;
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
    public static final Set<PrimitiveType.Primitive> INT_TYPES = Set.of(INT, LONG, BYTE, SHORT);

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    SchemaContributorBuildItem handleEnums(CombinedIndexBuildItem index, FTLRecorder recorder) {
        var enumAnnotations = index.getIndex().getAnnotations(FTLDotNames.ENUM);
        log.info("Processing {} enum annotations into decls", enumAnnotations.size());

        return new SchemaContributorBuildItem(new Consumer<ModuleBuilder>() {
            @Override
            public void accept(ModuleBuilder moduleBuilder) {
                try {
                    var decls = extractEnumDecls(index, enumAnnotations, recorder, moduleBuilder);
                    for (var decl : decls) {
                        moduleBuilder.addDecls(decl);
                    }
                } catch (ClassNotFoundException | NoSuchFieldException | IllegalAccessException e) {
                    throw new RuntimeException(e);
                }
            }
        });
    }

    /**
     * Extract all enums for this module, returning a Decl for each. Also registers the enums with the recorder, which
     * sets up Jackson serialization in the runtime.
     * ModuleBuilder.buildType is used, and has the side effect of adding child Decls to the module.
     */
    private List<Decl> extractEnumDecls(CombinedIndexBuildItem index, Collection<AnnotationInstance> enumAnnotations,
            FTLRecorder recorder, ModuleBuilder moduleBuilder)
            throws ClassNotFoundException, NoSuchFieldException, IllegalAccessException {
        List<Decl> decls = new ArrayList<>();
        for (var enumAnnotation : enumAnnotations) {
            boolean exported = enumAnnotation.target().hasAnnotation(FTLDotNames.EXPORT);
            ClassInfo classInfo = enumAnnotation.target().asClass();
            Class<?> clazz = Class.forName(classInfo.name().toString(), false,
                    Thread.currentThread().getContextClassLoader());
            var isLocalToModule = !classInfo.hasDeclaredAnnotation(GENERATED_REF);

            if (classInfo.isEnum()) {
                // Value enum
                recorder.registerEnum(clazz);
                if (isLocalToModule) {
                    decls.add(extractValueEnum(classInfo, clazz, exported));
                }
            } else {
                var typeEnum = extractTypeEnum(index, moduleBuilder, classInfo, exported);
                recorder.registerEnum(clazz, typeEnum.variantClasses);
                if (isLocalToModule) {
                    decls.add(typeEnum.decl);
                }
            }
        }
        return decls;
    }

    /**
     * Value enums are Java language enums with a single field 'value'
     */
    private Decl extractValueEnum(ClassInfo classInfo, Class<?> clazz, boolean exported)
            throws NoSuchFieldException, IllegalAccessException {
        Enum.Builder enumBuilder = Enum.newBuilder()
                .setName(classInfo.simpleName())
                .setExport(exported);
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

    private record TypeEnum(Decl decl, List<Class<?>> variantClasses) {
    }

    /**
     * Type Enums are an interface with 1+ implementing classes. The classes may be: </br>
     *  - a wrapper for a FTL native type e.g. string, [string]. Has @EnumHolder annotation </br>
     *  - a class with arbitrary fields </br>
     */
    private TypeEnum extractTypeEnum(CombinedIndexBuildItem index, ModuleBuilder moduleBuilder,
            ClassInfo classInfo, boolean exported) throws ClassNotFoundException {
        Enum.Builder enumBuilder = Enum.newBuilder()
                .setName(classInfo.simpleName())
                .setExport(exported);
        var variants = index.getComputingIndex().getAllKnownImplementors(classInfo.name());
        if (variants.isEmpty()) {
            throw new RuntimeException("No variants found for enum: " + enumBuilder.getName());
        }
        var variantClasses = new ArrayList<Class<?>>();
        for (var variant : variants) {
            Type variantType;
            if (variant.hasAnnotation(ENUM_HOLDER)) {
                // Enum value holder class
                FieldInfo valueField = variant.field("value");
                if (valueField == null) {
                    throw new RuntimeException("Enum variant must have a 'value' field: " + variant.name());
                }
                variantType = valueField.type();
                // TODO add to variantClasses; write serialization code for holder classes
            } else {
                // Class is the enum variant type
                variantType = ClassType.builder(variant.name()).build();
                Class<?> variantClazz = Class.forName(variantType.name().toString(), false,
                        Thread.currentThread().getContextClassLoader());
                variantClasses.add(variantClazz);
            }
            xyz.block.ftl.v1.schema.Type declType = moduleBuilder.buildType(variantType, exported,
                    Nullability.NOT_NULL);
            TypeValue typeValue = TypeValue.newBuilder().setValue(declType).build();

            EnumVariant.Builder variantBuilder = EnumVariant.newBuilder()
                    .setName(variant.simpleName())
                    .setValue(Value.newBuilder().setTypeValue(typeValue).build());
            enumBuilder.addVariants(variantBuilder.build());
        }
        return new TypeEnum(Decl.newBuilder().setEnum(enumBuilder).build(), variantClasses);
    }

    private boolean isInt(Type type) {
        return type.kind() == Type.Kind.PRIMITIVE && INT_TYPES.contains(type.asPrimitiveType().primitive());
    }

}
