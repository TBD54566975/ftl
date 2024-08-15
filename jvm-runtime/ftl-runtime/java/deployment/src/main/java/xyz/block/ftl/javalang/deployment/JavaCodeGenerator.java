package xyz.block.ftl.javalang.deployment;

import java.io.IOException;
import java.lang.annotation.Retention;
import java.nio.file.Path;
import java.time.ZonedDateTime;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeMap;

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

import xyz.block.ftl.GeneratedRef;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.VerbClientDefinition;
import xyz.block.ftl.VerbClientEmpty;
import xyz.block.ftl.VerbClientSink;
import xyz.block.ftl.VerbClientSource;
import xyz.block.ftl.deployment.JVMCodeGenerator;
import xyz.block.ftl.v1.schema.Data;
import xyz.block.ftl.v1.schema.Enum;
import xyz.block.ftl.v1.schema.Module;
import xyz.block.ftl.v1.schema.Topic;
import xyz.block.ftl.v1.schema.Type;
import xyz.block.ftl.v1.schema.Verb;

public class JavaCodeGenerator extends JVMCodeGenerator {

    public static final String CLIENT = "Client";
    public static final String PACKAGE_PREFIX = "ftl.";

    protected void generateTopicSubscription(Module module, Topic data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Path outputDir) throws IOException {
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

    protected void generateEnum(Module module, Enum data, String packageName, Map<DeclRef, Type> typeAliasMap, Path outputDir)
            throws IOException {
        String thisType = className(data.getName());
        TypeSpec.Builder dataBuilder = TypeSpec.enumBuilder(thisType)
                .addAnnotation(
                        AnnotationSpec.builder(GeneratedRef.class)
                                .addMember("name", "\"" + data.getName() + "\"")
                                .addMember("module", "\"" + module.getName() + "\"").build())
                .addModifiers(Modifier.PUBLIC);

        for (var i : data.getVariantsList()) {
            dataBuilder.addEnumConstant(i.getName());
        }

        JavaFile javaFile = JavaFile.builder(packageName, dataBuilder.build())
                .build();

        javaFile.writeTo(outputDir);
    }

    protected void generateDataObject(Module module, Data data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Path outputDir) throws IOException {
        String thisType = className(data.getName());
        TypeSpec.Builder dataBuilder = TypeSpec.classBuilder(thisType)
                .addAnnotation(
                        AnnotationSpec.builder(GeneratedRef.class)
                                .addMember("name", "\"" + data.getName() + "\"")
                                .addMember("module", "\"" + module.getName() + "\"").build())
                .addModifiers(Modifier.PUBLIC);
        MethodSpec.Builder allConstructor = MethodSpec.constructorBuilder().addModifiers(Modifier.PUBLIC);

        dataBuilder.addMethod(allConstructor.build());
        for (var param : data.getTypeParametersList()) {
            dataBuilder.addTypeVariable(TypeVariableName.get(param.getName()));
        }
        Map<String, Runnable> sortedFields = new TreeMap<>();

        for (var i : data.getFieldsList()) {
            TypeName dataType = toAnnotatedJavaTypeName(i.getType(), typeAliasMap);
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

    protected void generateVerb(Module module, Verb verb, String packageName, Map<DeclRef, Type> typeAliasMap, Path outputDir)
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
                    toJavaTypeName(verb.getResponse(), typeAliasMap, true)));
            typeBuilder.addMethod(MethodSpec.methodBuilder("call")
                    .returns(toAnnotatedJavaTypeName(verb.getResponse(), typeAliasMap))
                    .addModifiers(Modifier.ABSTRACT).addModifiers(Modifier.PUBLIC).build());
        } else if (verb.getResponse().hasUnit()) {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClientSink.class),
                    toJavaTypeName(verb.getRequest(), typeAliasMap, true)));
            typeBuilder.addMethod(MethodSpec.methodBuilder("call").returns(TypeName.VOID)
                    .addParameter(toAnnotatedJavaTypeName(verb.getRequest(), typeAliasMap), "value")
                    .addModifiers(Modifier.ABSTRACT).addModifiers(Modifier.PUBLIC).build());
        } else {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClient.class),
                    toJavaTypeName(verb.getRequest(), typeAliasMap, true),
                    toJavaTypeName(verb.getResponse(), typeAliasMap, true)));
            typeBuilder.addMethod(MethodSpec.methodBuilder("call")
                    .returns(toAnnotatedJavaTypeName(verb.getResponse(), typeAliasMap))
                    .addParameter(toAnnotatedJavaTypeName(verb.getRequest(), typeAliasMap), "value")
                    .addModifiers(Modifier.ABSTRACT).addModifiers(Modifier.PUBLIC).build());
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

    private TypeName toAnnotatedJavaTypeName(Type type, Map<DeclRef, Type> typeAliasMap) {
        var results = toJavaTypeName(type, typeAliasMap, false);
        if (type.hasRef() || type.hasArray() || type.hasBytes() || type.hasString() || type.hasMap() || type.hasTime()) {
            return results.annotated(AnnotationSpec.builder(NotNull.class).build());
        }
        return results;
    }

    private TypeName toJavaTypeName(Type type, Map<DeclRef, Type> typeAliasMap, boolean boxPrimitives) {
        if (type.hasArray()) {
            return ParameterizedTypeName.get(ClassName.get(List.class),
                    toJavaTypeName(type.getArray().getElement(), typeAliasMap, false));
        } else if (type.hasString()) {
            return ClassName.get(String.class);
        } else if (type.hasOptional()) {
            // Always box for optional, as normal primities can't be null
            return toJavaTypeName(type.getOptional().getType(), typeAliasMap, true);
        } else if (type.hasRef()) {
            if (type.getRef().getModule().isEmpty()) {
                return TypeVariableName.get(type.getRef().getName());
            }
            DeclRef key = new DeclRef(type.getRef().getModule(), type.getRef().getName());
            if (typeAliasMap.containsKey(key)) {
                return toJavaTypeName(typeAliasMap.get(key), typeAliasMap, boxPrimitives);
            }
            var params = type.getRef().getTypeParametersList();
            ClassName className = ClassName.get(PACKAGE_PREFIX + type.getRef().getModule(), type.getRef().getName());
            if (params.isEmpty()) {
                return className;
            }
            List<TypeName> javaTypes = params.stream()
                    .map(s -> s.hasUnit() ? WildcardTypeName.subtypeOf(Object.class) : toJavaTypeName(s, typeAliasMap, true))
                    .toList();
            return ParameterizedTypeName.get(className, javaTypes.toArray(new TypeName[javaTypes.size()]));
        } else if (type.hasMap()) {
            return ParameterizedTypeName.get(ClassName.get(Map.class),
                    toJavaTypeName(type.getMap().getKey(), typeAliasMap, true),
                    toJavaTypeName(type.getMap().getValue(), typeAliasMap, true));
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

    protected static final Set<String> JAVA_KEYWORDS = Set.of("abstract", "continue", "for", "new", "switch", "assert",
            "default", "goto", "package", "synchronized", "boolean", "do", "if", "private", "this", "break", "double",
            "implements", "protected", "throw", "byte", "else", "import", "public", "throws", "case", "enum", "instanceof",
            "return", "transient", "catch", "extends", "int", "short", "try", "char", "final", "interface", "static", "void",
            "class", "finally", "long", "strictfp", "volatile", "const", "float", "native", "super", "while");
}
