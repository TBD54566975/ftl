package xyz.block.ftl.javalang.deployment;

import java.io.IOException;
import java.lang.annotation.Retention;
import java.nio.file.Path;
import java.time.ZonedDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.TreeMap;
import java.util.stream.Collectors;

import javax.lang.model.element.Modifier;

import org.jetbrains.annotations.NotNull;

import com.squareup.javapoet.AnnotationSpec;
import com.squareup.javapoet.ArrayTypeName;
import com.squareup.javapoet.ClassName;
import com.squareup.javapoet.JavaFile;
import com.squareup.javapoet.MethodSpec;
import com.squareup.javapoet.ParameterizedTypeName;
import com.squareup.javapoet.TypeName;
import com.squareup.javapoet.TypeSpec;
import com.squareup.javapoet.TypeVariableName;
import com.squareup.javapoet.WildcardTypeName;

import xyz.block.ftl.EnumHolder;
import xyz.block.ftl.GeneratedRef;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.TypeAlias;
import xyz.block.ftl.TypeAliasMapper;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.VerbClientDefinition;
import xyz.block.ftl.VerbClientEmpty;
import xyz.block.ftl.VerbClientSink;
import xyz.block.ftl.VerbClientSource;
import xyz.block.ftl.deployment.JVMCodeGenerator;
import xyz.block.ftl.v1.schema.Data;
import xyz.block.ftl.v1.schema.Enum;
import xyz.block.ftl.v1.schema.EnumVariant;
import xyz.block.ftl.v1.schema.Module;
import xyz.block.ftl.v1.schema.Topic;
import xyz.block.ftl.v1.schema.Type;
import xyz.block.ftl.v1.schema.Value;
import xyz.block.ftl.v1.schema.Verb;

public class JavaCodeGenerator extends JVMCodeGenerator {

    public static final String CLIENT = "Client";
    public static final String PACKAGE_PREFIX = "ftl.";

    @Override
    protected void generateTypeAliasMapper(String module, String name, String packageName, Optional<String> nativeTypeAlias,
            Path outputDir) throws IOException {
        TypeSpec.Builder typeBuilder = TypeSpec.interfaceBuilder(className(name) + TYPE_MAPPER)
                .addAnnotation(AnnotationSpec.builder(TypeAlias.class)
                        .addMember("name", "\"" + name + "\"")
                        .addMember("module", "\"" + module + "\"")
                        .build())
                .addModifiers(Modifier.PUBLIC);
        if (nativeTypeAlias.isEmpty()) {
            TypeVariableName finalType = TypeVariableName.get("T");
            typeBuilder.addTypeVariable(finalType);
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(TypeAliasMapper.class),
                    finalType, ClassName.get(String.class)));
        } else {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(TypeAliasMapper.class),
                    ClassName.bestGuess(nativeTypeAlias.get()), ClassName.get(String.class)));
        }
        TypeSpec theType = typeBuilder
                .build();

        JavaFile javaFile = JavaFile.builder(packageName, theType)
                .build();

        javaFile.writeTo(outputDir);
    }

    protected void generateTopicSubscription(Module module, Topic data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Path outputDir) throws IOException {
        String thisType = className(data.getName() + "Subscription");

        TypeSpec.Builder dataBuilder = TypeSpec.annotationBuilder(thisType)
                .addModifiers(Modifier.PUBLIC);
        if (data.getEvent().hasRef()) {
            dataBuilder.addJavadoc("Subscription to the topic of type {@link $L}",
                    data.getEvent().getRef().getName());
        }
        dataBuilder.addAnnotation(AnnotationSpec.builder(Retention.class)
                .addMember("value", "java.lang.annotation.RetentionPolicy.RUNTIME").build());
        dataBuilder.addAnnotation(AnnotationSpec.builder(Subscription.class)
                .addMember("topic", "\"" + data.getName() + "\"")
                .addMember("module", "\"" + module.getName() + "\"")
                .addMember("name", "\"" + data.getName() + "Subscription\"")
                .build());

        JavaFile javaFile = JavaFile.builder(packageName, dataBuilder.build())
                .build();

        javaFile.writeTo(outputDir);
    }

    protected void generateEnum(Module module, Enum data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Map<DeclRef, List<EnumInfo>> enumVariantInfoMap, Path outputDir)
            throws IOException {
        String interfaceType = className(data.getName());
        if (data.hasType()) {
            //Enums with a type are "value enums" - Java natively supports these
            TypeSpec.Builder dataBuilder = TypeSpec.enumBuilder(interfaceType)
                    .addAnnotation(getGeneratedRefAnnotation(module.getName(), data.getName()))
                    .addAnnotation(AnnotationSpec.builder(xyz.block.ftl.Enum.class).build())
                    .addModifiers(Modifier.PUBLIC);

            TypeName enumType = toAnnotatedJavaTypeName(data.getType(), typeAliasMap, nativeTypeAliasMap);
            dataBuilder.addField(enumType, "value", Modifier.PRIVATE, Modifier.FINAL);
            dataBuilder.addMethod(MethodSpec.constructorBuilder()
                    .addParameter(enumType, "value")
                    .addStatement("this.value = value")
                    .build());
            dataBuilder.addMethod(MethodSpec.methodBuilder("getValue")
                    .addModifiers(Modifier.PUBLIC)
                    .returns(enumType)
                    .addStatement("return value")
                    .build());

            var format = data.getType().hasString() ? "$S" : "$L";
            for (var i : data.getVariantsList()) {
                Object value = toJavaValue(i.getValue());
                dataBuilder.addEnumConstant(i.getName(), TypeSpec.anonymousClassBuilder(format, value).build());
            }
            JavaFile javaFile = JavaFile.builder(packageName, dataBuilder.build())
                    .build();
            javaFile.writeTo(outputDir);
        } else {
            // Enums without a type are (confusingly) "type enums". Java can't represent these directly, so we use a
            // sealed class

            // TODO JavaPoet doesn't support 'sealed' or 'permits' syntax yet, so we can't seal the interface
            // https://github.com/square/javapoet/issues/823
            TypeSpec.Builder interfaceBuilder = TypeSpec.interfaceBuilder(interfaceType)
                    .addAnnotation(getGeneratedRefAnnotation(module.getName(), data.getName()))
                    .addAnnotation(AnnotationSpec.builder(xyz.block.ftl.Enum.class).build())
                    .addModifiers(Modifier.PUBLIC);

            Map<String, TypeName> variantValuesTypes = data.getVariantsList().stream().collect(
                    Collectors.toMap(EnumVariant::getName, v -> toAnnotatedJavaTypeName(v.getValue().getTypeValue().getValue(),
                            typeAliasMap, nativeTypeAliasMap)));
            for (var variant : data.getVariantsList()) {
                // Interface has isX and getX methods for each variant
                String name = variant.getName();
                TypeName valueTypeName = variantValuesTypes.get(name);
                interfaceBuilder.addMethod(MethodSpec.methodBuilder("is" + name)
                        .addModifiers(Modifier.PUBLIC, Modifier.ABSTRACT)
                        .returns(TypeName.BOOLEAN)
                        .build());
                interfaceBuilder.addMethod(MethodSpec.methodBuilder("get" + name)
                        .addModifiers(Modifier.PUBLIC, Modifier.ABSTRACT)
                        .returns(valueTypeName)
                        .build());

                if (variant.getValue().getTypeValue().getValue().hasRef()) {
                    // Value type is a Ref, so it will have a class generated by generateDataObject
                    // Store this variant in enumVariantInfoMap so we can fetch it later
                    DeclRef key = new DeclRef(module.getName(), name);
                    List<EnumInfo> variantInfos = enumVariantInfoMap.computeIfAbsent(key, k -> new ArrayList<>());
                    variantInfos.add(new EnumInfo(interfaceType, variant, data.getVariantsList()));
                } else {
                    // Value type isn't a Ref, so we make a wrapper class that implements our interface
                    TypeSpec.Builder dataBuilder = TypeSpec.classBuilder(className(name))
                            .addAnnotation(getGeneratedRefAnnotation(module.getName(), name))
                            .addAnnotation(AnnotationSpec.builder(EnumHolder.class).build())
                            .addModifiers(Modifier.PUBLIC, Modifier.FINAL);
                    dataBuilder.addField(valueTypeName, "value", Modifier.PRIVATE, Modifier.FINAL);
                    dataBuilder.addMethod(MethodSpec.constructorBuilder()
                            .addStatement("this.value = null")
                            .addModifiers(Modifier.PRIVATE)
                            .build());
                    dataBuilder.addMethod(MethodSpec.constructorBuilder()
                            .addParameter(valueTypeName, "value")
                            .addStatement("this.value = value")
                            .addModifiers(Modifier.PUBLIC)
                            .build());
                    addTypeEnumInterfaceMethods(packageName, interfaceType, dataBuilder, name, valueTypeName,
                            variantValuesTypes, false);
                    JavaFile javaFile = JavaFile.builder(packageName, dataBuilder.build())
                            .build();
                    javaFile.writeTo(outputDir);
                }
                JavaFile javaFile = JavaFile.builder(packageName, interfaceBuilder.build())
                        .build();
                javaFile.writeTo(outputDir);
            }
        }
    }

    protected void generateDataObject(Module module, Data data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Map<DeclRef, List<EnumInfo>> enumVariantInfoMap, Path outputDir)
            throws IOException {
        String thisType = className(data.getName());
        TypeSpec.Builder dataBuilder = TypeSpec.classBuilder(thisType)
                .addAnnotation(getGeneratedRefAnnotation(module.getName(), data.getName()))
                .addModifiers(Modifier.PUBLIC);

        // if data is part of a type enum, generate the interface methods for each variant
        DeclRef key = new DeclRef(module.getName(), data.getName());
        if (enumVariantInfoMap.containsKey(key)) {
            for (var enumVariantInfo : enumVariantInfoMap.get(key)) {
                String name = enumVariantInfo.variant().getName();
                TypeName variantTypeName = ClassName.get(packageName, name);
                Map<String, TypeName> variantValuesTypes = enumVariantInfo.otherVariants().stream().collect(
                        Collectors.toMap(EnumVariant::getName,
                                v -> toAnnotatedJavaTypeName(v.getValue().getTypeValue().getValue(),
                                        typeAliasMap, nativeTypeAliasMap)));
                addTypeEnumInterfaceMethods(packageName, enumVariantInfo.interfaceType(), dataBuilder, name,
                        variantTypeName, variantValuesTypes, true);
            }
        }

        MethodSpec.Builder allConstructor = MethodSpec.constructorBuilder().addModifiers(Modifier.PUBLIC);
        dataBuilder.addMethod(allConstructor.build());
        for (var param : data.getTypeParametersList()) {
            dataBuilder.addTypeVariable(TypeVariableName.get(param.getName()));
        }
        Map<String, Runnable> sortedFields = new TreeMap<>();

        for (var i : data.getFieldsList()) {
            TypeName dataType = toAnnotatedJavaTypeName(i.getType(), typeAliasMap, nativeTypeAliasMap);
            String name = i.getName();
            var fieldName = toJavaName(name);
            dataBuilder.addField(dataType, fieldName, Modifier.PRIVATE);
            sortedFields.put(fieldName, () -> {
                allConstructor.addParameter(dataType, fieldName);
                allConstructor.addCode("this.$L = $L;\n", fieldName, fieldName);
            });
            String methodName = Character.toUpperCase(name.charAt(0)) + name.substring(1);
            dataBuilder.addMethod(MethodSpec.methodBuilder("set" + methodName)
                    .addModifiers(Modifier.PUBLIC)
                    .addParameter(dataType, fieldName)
                    .returns(ClassName.get(packageName, thisType))
                    .addCode("this.$L = $L;\n", fieldName, fieldName)
                    .addCode("return this;")
                    .build());
            if (i.getType().hasBool()) {
                dataBuilder.addMethod(MethodSpec.methodBuilder("is" + methodName)
                        .addModifiers(Modifier.PUBLIC)
                        .returns(dataType)
                        .addCode("return $L;", fieldName)
                        .build());
            } else {
                dataBuilder.addMethod(MethodSpec.methodBuilder("get" + methodName)
                        .addModifiers(Modifier.PUBLIC)
                        .returns(dataType)
                        .addCode("return $L;", fieldName)
                        .build());
            }
        }
        if (!sortedFields.isEmpty()) {

            for (var v : sortedFields.values()) {
                v.run();
            }
            dataBuilder.addMethod(allConstructor.build());

        }
        JavaFile javaFile = JavaFile.builder(packageName, dataBuilder.build())
                .build();

        javaFile.writeTo(outputDir);
    }

    protected void generateVerb(Module module, Verb verb, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Path outputDir)
            throws IOException {
        TypeSpec.Builder typeBuilder = TypeSpec.interfaceBuilder(className(verb.getName()) + CLIENT)
                .addAnnotation(AnnotationSpec.builder(VerbClientDefinition.class)
                        .addMember("name", "\"" + verb.getName() + "\"")
                        .addMember("module", "\"" + module.getName() + "\"")
                        .build())
                .addModifiers(Modifier.PUBLIC);
        if (verb.getRequest().hasUnit() && verb.getResponse().hasUnit()) {
            typeBuilder.addSuperinterface(ClassName.get(VerbClientEmpty.class));
        } else if (verb.getRequest().hasUnit()) {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClientSource.class),
                    toJavaTypeName(verb.getResponse(), typeAliasMap, nativeTypeAliasMap, true)));
            typeBuilder.addMethod(MethodSpec.methodBuilder("call")
                    .returns(toAnnotatedJavaTypeName(verb.getResponse(), typeAliasMap, nativeTypeAliasMap))
                    .addModifiers(Modifier.ABSTRACT, Modifier.PUBLIC).build());
        } else if (verb.getResponse().hasUnit()) {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClientSink.class),
                    toJavaTypeName(verb.getRequest(), typeAliasMap, nativeTypeAliasMap, true)));
            typeBuilder.addMethod(MethodSpec.methodBuilder("call").returns(TypeName.VOID)
                    .addParameter(toAnnotatedJavaTypeName(verb.getRequest(), typeAliasMap, nativeTypeAliasMap), "value")
                    .addModifiers(Modifier.ABSTRACT, Modifier.PUBLIC).build());
        } else {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClient.class),
                    toJavaTypeName(verb.getRequest(), typeAliasMap, nativeTypeAliasMap, true),
                    toJavaTypeName(verb.getResponse(), typeAliasMap, nativeTypeAliasMap, true)));
            typeBuilder.addMethod(MethodSpec.methodBuilder("call")
                    .returns(toAnnotatedJavaTypeName(verb.getResponse(), typeAliasMap, nativeTypeAliasMap))
                    .addParameter(toAnnotatedJavaTypeName(verb.getRequest(), typeAliasMap, nativeTypeAliasMap), "value")
                    .addModifiers(Modifier.ABSTRACT, Modifier.PUBLIC).build());
        }

        TypeSpec helloWorld = typeBuilder
                .build();

        JavaFile javaFile = JavaFile.builder(packageName, helloWorld)
                .build();

        javaFile.writeTo(outputDir);
    }

    private String toJavaName(String name) {
        if (JAVA_KEYWORDS.contains(name)) {
            return name + "_";
        }
        return name;
    }

    private TypeName toAnnotatedJavaTypeName(Type type, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap) {
        var results = toJavaTypeName(type, typeAliasMap, nativeTypeAliasMap, false);
        if (type.hasRef() || type.hasArray() || type.hasBytes() || type.hasString() || type.hasMap() || type.hasTime()) {
            return results.annotated(AnnotationSpec.builder(NotNull.class).build());
        }
        return results;
    }

    private TypeName toJavaTypeName(Type type, Map<DeclRef, Type> typeAliasMap, Map<DeclRef, String> nativeTypeAliasMap,
            boolean boxPrimitives) {
        if (type.hasArray()) {
            return ParameterizedTypeName.get(ClassName.get(List.class),
                    toJavaTypeName(type.getArray().getElement(), typeAliasMap, nativeTypeAliasMap, false));
        } else if (type.hasString()) {
            return ClassName.get(String.class);
        } else if (type.hasOptional()) {
            // Always box for optional, as normal primitives can't be null
            return toJavaTypeName(type.getOptional().getType(), typeAliasMap, nativeTypeAliasMap, true);
        } else if (type.hasRef()) {
            if (type.getRef().getModule().isEmpty()) {
                return TypeVariableName.get(type.getRef().getName());
            }
            DeclRef key = new DeclRef(type.getRef().getModule(), type.getRef().getName());
            if (nativeTypeAliasMap.containsKey(key)) {
                return ClassName.bestGuess(nativeTypeAliasMap.get(key));
            }
            if (typeAliasMap.containsKey(key)) {
                return toJavaTypeName(typeAliasMap.get(key), typeAliasMap, nativeTypeAliasMap, boxPrimitives);
            }
            var params = type.getRef().getTypeParametersList();
            ClassName className = ClassName.get(PACKAGE_PREFIX + type.getRef().getModule(), type.getRef().getName());
            if (params.isEmpty()) {
                return className;
            }
            List<TypeName> javaTypes = params.stream()
                    .map(s -> s.hasUnit() ? WildcardTypeName.subtypeOf(Object.class)
                            : toJavaTypeName(s, typeAliasMap, nativeTypeAliasMap, true))
                    .toList();
            return ParameterizedTypeName.get(className, javaTypes.toArray(new TypeName[javaTypes.size()]));
        } else if (type.hasMap()) {
            return ParameterizedTypeName.get(ClassName.get(Map.class),
                    toJavaTypeName(type.getMap().getKey(), typeAliasMap, nativeTypeAliasMap, true),
                    toJavaTypeName(type.getMap().getValue(), typeAliasMap, nativeTypeAliasMap, true));
        } else if (type.hasTime()) {
            return ClassName.get(ZonedDateTime.class);
        } else if (type.hasInt()) {
            return boxPrimitives ? ClassName.get(Long.class) : TypeName.LONG;
        } else if (type.hasUnit()) {
            return TypeName.VOID;
        } else if (type.hasBool()) {
            return boxPrimitives ? ClassName.get(Boolean.class) : TypeName.BOOLEAN;
        } else if (type.hasFloat()) {
            return boxPrimitives ? ClassName.get(Double.class) : TypeName.DOUBLE;
        } else if (type.hasBytes()) {
            return ArrayTypeName.of(TypeName.BYTE);
        } else if (type.hasAny()) {
            return TypeName.OBJECT;
        }

        throw new RuntimeException("Cannot generate Java type name: " + type);
    }

    /**
     * Get concrete value from a Value
     */
    private Object toJavaValue(Value value) {
        if (value.hasIntValue()) {
            return value.getIntValue().getValue();
        } else if (value.hasStringValue()) {
            return value.getStringValue().getValue();
        } else if (value.hasTypeValue()) {
            // Can't instantiate a TypeValue now. Cannot happen because it's only used in type enums
            throw new RuntimeException("Cannot generate TypeValue: " + value);
        }
        throw new RuntimeException("Cannot generate Java value: " + value);
    }

    /**
     * Adds the super interface and isX, getX methods to the <code>dataBuilder</code> for a type enum variant
     */
    private static void addTypeEnumInterfaceMethods(String packageName, String interfaceType, TypeSpec.Builder dataBuilder,
            String enumVariantName, TypeName variantTypeName, Map<String, TypeName> variantValuesTypes, boolean returnSelf) {

        dataBuilder.addSuperinterface(ClassName.get(packageName, interfaceType));

        // Positive implementation of isX, getX for its type
        dataBuilder.addMethod(MethodSpec.methodBuilder("is" + enumVariantName)
                .addModifiers(Modifier.PUBLIC)
                .returns(TypeName.BOOLEAN)
                .addStatement("return true")
                .build());

        MethodSpec.Builder getMethod = MethodSpec.methodBuilder("get" + enumVariantName)
                .addModifiers(Modifier.PUBLIC)
                .returns(variantTypeName);
        if (returnSelf) {
            getMethod.addStatement("return this");
        } else {
            getMethod.addStatement("return value");
        }
        dataBuilder.addMethod(getMethod.build());

        for (var thingIAmNot : variantValuesTypes.entrySet()) {
            if (thingIAmNot.getKey().equals(enumVariantName)) {
                continue;
            }
            // Negative implementation of isX, getX for other types
            dataBuilder.addMethod(MethodSpec.methodBuilder("is" + thingIAmNot.getKey())
                    .addModifiers(Modifier.PUBLIC)
                    .returns(TypeName.BOOLEAN)
                    .addStatement("return false")
                    .build());
            dataBuilder.addMethod(MethodSpec.methodBuilder("get" + thingIAmNot.getKey())
                    .addModifiers(Modifier.PUBLIC)
                    .returns(thingIAmNot.getValue())
                    .addStatement("throw new UnsupportedOperationException()")
                    .build());
        }
    }

    private static @NotNull AnnotationSpec getGeneratedRefAnnotation(String module, String name) {
        return AnnotationSpec.builder(GeneratedRef.class)
                .addMember("name", "\"" + name + "\"")
                .addMember("module", "\"" + module + "\"").build();
    }

    protected static final Set<String> JAVA_KEYWORDS = Set.of("abstract", "continue", "for", "new", "switch", "assert",
            "default", "goto", "package", "synchronized", "boolean", "do", "if", "private", "this", "break", "double",
            "implements", "protected", "throw", "byte", "else", "import", "public", "throws", "case", "enum", "instanceof",
            "return", "transient", "catch", "extends", "int", "short", "try", "char", "final", "interface", "static", "void",
            "class", "finally", "long", "strictfp", "volatile", "const", "float", "native", "super", "while");
}
