package xyz.block.ftl.kotlin.deployment;

import java.io.IOException;
import java.nio.file.Path;
import java.time.ZonedDateTime;
import java.util.List;
import java.util.Map;
import java.util.Set;

import com.squareup.kotlinpoet.AnnotationSpec;
import com.squareup.kotlinpoet.ClassName;
import com.squareup.kotlinpoet.CodeBlock;
import com.squareup.kotlinpoet.FileSpec;
import com.squareup.kotlinpoet.FunSpec;
import com.squareup.kotlinpoet.KModifier;
import com.squareup.kotlinpoet.ParameterizedTypeName;
import com.squareup.kotlinpoet.PropertySpec;
import com.squareup.kotlinpoet.TypeName;
import com.squareup.kotlinpoet.TypeSpec;
import com.squareup.kotlinpoet.TypeVariableName;
import com.squareup.kotlinpoet.WildcardTypeName;

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

public class KotlinCodeGenerator extends JVMCodeGenerator {

    public static final String CLIENT = "Client";
    public static final String PACKAGE_PREFIX = "ftl.";

    protected void generateTopicSubscription(Module module, Topic data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Path outputDir) throws IOException {
        String thisType = className(data.getName() + "Subscription");

        TypeSpec.Builder dataBuilder = TypeSpec.annotationBuilder(ClassName.bestGuess(thisType));
        dataBuilder.addModifiers(KModifier.PUBLIC);
        if (data.getEvent().hasRef()) {
            dataBuilder.addKdoc("Subscription to the topic of type {@link $L}",
                    data.getEvent().getRef().getName());
        }
        dataBuilder.addAnnotation(AnnotationSpec.builder(Subscription.class)
                .addMember("topic=\"" + data.getName() + "\"")
                .addMember("module=\"" + module.getName() + "\"")
                .addMember("name=\"" + data.getName() + "Subscription\"")
                .build());

        FileSpec javaFile = FileSpec.builder(packageName, thisType)
                .addType(dataBuilder.build())
                .build();

        javaFile.writeTo(outputDir);
    }

    protected void generateEnum(Module module, Enum data, String packageName, Map<DeclRef, Type> typeAliasMap, Path outputDir)
            throws IOException {
        String thisType = className(data.getName());
        TypeSpec.Builder dataBuilder = TypeSpec.enumBuilder(thisType)
                .addAnnotation(
                        AnnotationSpec.builder(GeneratedRef.class)
                                .addMember("name=\"" + data.getName() + "\"")
                                .addMember("module=\"" + module.getName() + "\"").build())
                .addModifiers(KModifier.PUBLIC);

        for (var i : data.getVariantsList()) {
            dataBuilder.addEnumConstant(i.getName());
        }

        FileSpec javaFile = FileSpec.builder(packageName, thisType)
                .addType(dataBuilder.build())
                .build();
        javaFile.writeTo(outputDir);
    }

    protected void generateDataObject(Module module, Data data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Path outputDir) throws IOException {
        String thisType = className(data.getName());
        TypeSpec.Builder dataBuilder = TypeSpec.classBuilder(thisType)
                .addAnnotation(
                        AnnotationSpec.builder(GeneratedRef.class)
                                .addMember("name=\"" + data.getName() + "\"")
                                .addMember("module=\"" + module.getName() + "\"").build())
                .addModifiers(KModifier.PUBLIC);
        if (!data.getFieldsList().isEmpty()) {
            dataBuilder.addModifiers(KModifier.DATA);
        }

        for (var param : data.getTypeParametersList()) {
            dataBuilder.addTypeVariable(TypeVariableName.get(param.getName()));
        }
        FunSpec.Builder constructorBuilder = FunSpec.constructorBuilder();

        for (var i : data.getFieldsList()) {
            TypeName dataType = toKotlinTypeName(i.getType(), typeAliasMap);
            String name = i.getName();
            var fieldName = toJavaName(name);
            constructorBuilder.addParameter(fieldName, dataType);
            dataBuilder.addProperty(PropertySpec.builder(fieldName, dataType, KModifier.PUBLIC)
                    .initializer(fieldName).build());

        }
        dataBuilder.primaryConstructor(constructorBuilder.build());
        FileSpec javaFile = FileSpec.builder(packageName, thisType)
                .addType(dataBuilder.build())
                .build();

        javaFile.writeTo(outputDir);
    }

    protected void generateVerb(Module module, Verb verb, String packageName, Map<DeclRef, Type> typeAliasMap, Path outputDir)
            throws IOException {
        String thisType = className(verb.getName()) + CLIENT;
        TypeSpec.Builder typeBuilder = TypeSpec.interfaceBuilder(thisType)
                .addAnnotation(AnnotationSpec.builder(VerbClientDefinition.class)
                        .addMember("name=\"" + verb.getName() + "\"")
                        .addMember("module=\"" + module.getName() + "\"")
                        .build())
                .addModifiers(KModifier.PUBLIC);
        if (verb.getRequest().hasUnit() && verb.getResponse().hasUnit()) {
            typeBuilder.addSuperinterface(className(VerbClientEmpty.class), CodeBlock.of(""));
        } else if (verb.getRequest().hasUnit()) {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(className(VerbClientSource.class),
                    toKotlinTypeName(verb.getResponse(), typeAliasMap)), CodeBlock.of(""));
            typeBuilder.addFunction(FunSpec.builder("call")
                    .returns(toKotlinTypeName(verb.getResponse(), typeAliasMap))
                    .addModifiers(KModifier.PUBLIC, KModifier.OVERRIDE, KModifier.ABSTRACT).build());
        } else if (verb.getResponse().hasUnit()) {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(className(VerbClientSink.class),
                    toKotlinTypeName(verb.getRequest(), typeAliasMap)), CodeBlock.of(""));
            typeBuilder.addFunction(FunSpec.builder("call")
                    .addModifiers(KModifier.OVERRIDE, KModifier.ABSTRACT)
                    .addParameter("value", toKotlinTypeName(verb.getRequest(), typeAliasMap)).build());
        } else {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(className(VerbClient.class),
                    toKotlinTypeName(verb.getRequest(), typeAliasMap),
                    toKotlinTypeName(verb.getResponse(), typeAliasMap)), CodeBlock.of(""));
            typeBuilder.addFunction(FunSpec.builder("call")
                    .returns(toKotlinTypeName(verb.getResponse(), typeAliasMap))
                    .addParameter("value", toKotlinTypeName(verb.getRequest(), typeAliasMap))
                    .addModifiers(KModifier.PUBLIC, KModifier.OVERRIDE, KModifier.ABSTRACT).build());
        }

        FileSpec javaFile = FileSpec.builder(packageName, thisType)
                .addType(typeBuilder.build())
                .build();

        javaFile.writeTo(outputDir);
    }

    private String toJavaName(String name) {
        if (JAVA_KEYWORDS.contains(name)) {
            return name + "_";
        }
        return name;
    }

    private ClassName className(Class<?> clazz) {
        if (clazz.getEnclosingClass() != null) {
            return className(clazz.getEnclosingClass()).nestedClass(clazz.getSimpleName());
        }
        return new ClassName(clazz.getPackage().getName(), clazz.getSimpleName());
    }

    private TypeName toKotlinTypeName(Type type, Map<DeclRef, Type> typeAliasMap) {
        if (type.hasArray()) {
            return ParameterizedTypeName.get(new ClassName("kotlin.collections", "List"),
                    toKotlinTypeName(type.getArray().getElement(), typeAliasMap));
        } else if (type.hasString()) {
            return new ClassName("kotlin", "String");
        } else if (type.hasOptional()) {
            // Always box for optional, as normal primities can't be null
            return toKotlinTypeName(type.getOptional().getType(), typeAliasMap);
        } else if (type.hasRef()) {
            if (type.getRef().getModule().isEmpty()) {
                return TypeVariableName.get(type.getRef().getName());
            }
            DeclRef key = new DeclRef(type.getRef().getModule(), type.getRef().getName());
            if (typeAliasMap.containsKey(key)) {
                return toKotlinTypeName(typeAliasMap.get(key), typeAliasMap);
            }
            var params = type.getRef().getTypeParametersList();
            ClassName className = new ClassName(PACKAGE_PREFIX + type.getRef().getModule(), type.getRef().getName());
            if (params.isEmpty()) {
                return className;
            }
            List<TypeName> javaTypes = params.stream()
                    .map(s -> s.hasUnit() ? WildcardTypeName.consumerOf(new ClassName("kotlin", "Any"))
                            : toKotlinTypeName(s, typeAliasMap))
                    .toList();
            return ParameterizedTypeName.get(className, javaTypes.toArray(new TypeName[javaTypes.size()]));
        } else if (type.hasMap()) {
            return ParameterizedTypeName.get(new ClassName("kotlin.collections", "Map"),
                    toKotlinTypeName(type.getMap().getKey(), typeAliasMap),
                    toKotlinTypeName(type.getMap().getValue(), typeAliasMap));
        } else if (type.hasTime()) {
            return className(ZonedDateTime.class);
        } else if (type.hasInt()) {
            return new ClassName("kotlin", "Long");
        } else if (type.hasUnit()) {
            return new ClassName("kotlin", "Unit");
        } else if (type.hasBool()) {
            return new ClassName("kotlin", "Boolean");
        } else if (type.hasFloat()) {
            return new ClassName("kotlin", "Double");
        } else if (type.hasBytes()) {
            return new ClassName("kotlin", "ByteArray");
        } else if (type.hasAny()) {
            return new ClassName("kotlin", "Any");
        }

        throw new RuntimeException("Cannot generate Kotlin type name: " + type);
    }

    // TODO: fix keywords
    protected static final Set<String> JAVA_KEYWORDS = Set.of("abstract", "continue", "for", "new", "switch", "assert",
            "default", "goto", "package", "synchronized", "boolean", "do", "if", "private", "this", "break", "double",
            "implements", "protected", "throw", "byte", "else", "import", "public", "throws", "case", "enum", "instanceof",
            "return", "transient", "catch", "extends", "int", "short", "try", "char", "final", "interface", "static", "void",
            "class", "finally", "long", "strictfp", "volatile", "const", "float", "native", "super", "while");
}
