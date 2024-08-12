package xyz.block.ftl.deployment;

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
import io.quarkus.bootstrap.model.ApplicationModel;
import io.quarkus.bootstrap.prebuild.CodeGenException;
import io.quarkus.deployment.CodeGenContext;
import io.quarkus.deployment.CodeGenProvider;
import org.eclipse.microprofile.config.Config;
import org.jboss.logging.Logger;
import org.jetbrains.annotations.NotNull;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.VerbClientDefinition;
import xyz.block.ftl.VerbClientEmpty;
import xyz.block.ftl.VerbClientSink;
import xyz.block.ftl.VerbClientSource;
import xyz.block.ftl.v1.schema.Module;
import xyz.block.ftl.v1.schema.Type;

import javax.lang.model.element.Modifier;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Instant;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeMap;
import java.util.stream.Stream;

public class FTLCodeGenerator implements CodeGenProvider {

    private static final Logger log = Logger.getLogger(FTLCodeGenerator.class);

    public static final String CLIENT = "Client";
    public static final String PACKAGE_PREFIX = "ftl.";
    String moduleName;

    @Override
    public void init(ApplicationModel model, Map<String, String> properties) {
        CodeGenProvider.super.init(model, properties);
        moduleName = model.getAppArtifact().getArtifactId();
    }

    @Override
    public String providerId() {
        return "ftl-clients";
    }

    @Override
    public String inputDirectory() {
        return "ftl-module-schema";
    }

    @Override
    public boolean trigger(CodeGenContext context) throws CodeGenException {
        if (!Files.isDirectory(context.inputDir())) {
            return false;
        }

        List<Module> modules = new ArrayList<>();

        Map<Key, Type> typeAliasMap = new HashMap<>();

        try (Stream<Path> pathStream = Files.list(context.inputDir())) {
            for (var file : pathStream.toList()) {
                String fileName = file.getFileName().toString();
                if (!fileName.endsWith(".pb")) {
                    continue;
                }
                var module = Module.parseFrom(Files.readAllBytes(file));
                for (var decl : module.getDeclsList()) {
                    if (decl.hasTypeAlias()) {
                        var data = decl.getTypeAlias();
                        typeAliasMap.put(new Key(module.getName(), data.getName()), data.getType());
                    }
                }
                modules.add(module);
            }
        } catch (IOException e) {
            throw new CodeGenException(e);
        }
        try {
            for (var module : modules) {
                String packageName = PACKAGE_PREFIX + module.getName();
                for (var decl : module.getDeclsList()) {
                    if (decl.hasVerb()) {
                        var verb = decl.getVerb();
                        if (!verb.getExport()) {
                            continue;
                        }

                        TypeSpec.Builder typeBuilder = TypeSpec.interfaceBuilder(className(verb.getName()) + CLIENT)
                                .addAnnotation(AnnotationSpec.builder(VerbClientDefinition.class)
                                        .addMember("name", "\"" + verb.getName() + "\"")
                                        .addMember("module", "\"" + module.getName() + "\"")
                                        .build())
                                .addModifiers(Modifier.PUBLIC);
                        if (verb.getRequest().hasUnit() && verb.getResponse().hasUnit()) {
                            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClientEmpty.class)));
                        } else if (verb.getRequest().hasUnit()) {
                            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClientSource.class), toJavaTypeName(verb.getResponse(), typeAliasMap)));
                            typeBuilder.addMethod(MethodSpec.methodBuilder("call").returns(toAnnotatedJavaTypeName(verb.getResponse(), typeAliasMap)).addModifiers(Modifier.ABSTRACT).addModifiers(Modifier.PUBLIC).build());
                        } else if (verb.getResponse().hasUnit()) {
                            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClientSink.class), toJavaTypeName(verb.getRequest(), typeAliasMap)));
                            typeBuilder.addMethod(MethodSpec.methodBuilder("call").returns(TypeName.VOID).addParameter(toAnnotatedJavaTypeName(verb.getRequest(), typeAliasMap), "value").addModifiers(Modifier.ABSTRACT).addModifiers(Modifier.PUBLIC).build());
                        } else {
                            typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClient.class), toJavaTypeName(verb.getRequest(), typeAliasMap), toJavaTypeName(verb.getResponse(), typeAliasMap)));
                            typeBuilder.addMethod(MethodSpec.methodBuilder("call").returns(toAnnotatedJavaTypeName(verb.getResponse(), typeAliasMap)).addParameter(toAnnotatedJavaTypeName(verb.getRequest(), typeAliasMap), "value").addModifiers(Modifier.ABSTRACT).addModifiers(Modifier.PUBLIC).build());
                        }

                        TypeSpec helloWorld = typeBuilder
                                .build();

                        JavaFile javaFile = JavaFile.builder(packageName, helloWorld)
                                .build();

                        javaFile.writeTo(context.outDir());

                    } else if (decl.hasData()) {
                        var data = decl.getData();
                        String thisType = className(data.getName());
                        TypeSpec.Builder dataBuilder = TypeSpec.classBuilder(thisType)
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

                        javaFile.writeTo(context.outDir());


                    } else if (decl.hasEnum()) {
                        var data = decl.getEnum();
                        String thisType = className(data.getName());
                        TypeSpec.Builder dataBuilder = TypeSpec.enumBuilder(thisType)
                                .addModifiers(Modifier.PUBLIC);

                        for (var i : data.getVariantsList()) {
                            dataBuilder.addEnumConstant(i.getName());
                        }

                        JavaFile javaFile = JavaFile.builder(packageName, dataBuilder.build())
                                .build();

                        javaFile.writeTo(context.outDir());

                    }
                }
            }

        } catch (Exception e) {
            throw new CodeGenException(e);
        }
        return true;
    }

    private String toJavaName(String name) {
        if (JAVA_KEYWORDS.contains(name)) {
            return name + "_";
        }
        return name;
    }

    private TypeName toAnnotatedJavaTypeName(Type type, Map<Key, Type> typeAliasMap) {
        var results = toJavaTypeName(type, typeAliasMap);
        if (type.hasRef() || type.hasArray() || type.hasBytes() || type.hasString() || type.hasMap() || type.hasTime()) {
            return results.annotated(AnnotationSpec.builder(NotNull.class).build());
        }
        return results;
    }

    private TypeName toJavaTypeName(Type type, Map<Key, Type> typeAliasMap) {
        if (type.hasArray()) {
            return ParameterizedTypeName.get(ClassName.get(List.class), toJavaTypeName(type.getArray().getElement(), typeAliasMap));
        } else if (type.hasString()) {
            return ClassName.get(String.class);
        } else if (type.hasOptional()) {
            return toJavaTypeName(type.getOptional().getType(), typeAliasMap);
        } else if (type.hasRef()) {
            if (type.getRef().getModule().isEmpty()) {
                return TypeVariableName.get(type.getRef().getName());
            }

            Key key = new Key(type.getRef().getModule(), type.getRef().getName());
            if (typeAliasMap.containsKey(key)) {
                return toJavaTypeName(typeAliasMap.get(key), typeAliasMap);
            }
            var params = type.getRef().getTypeParametersList();
            ClassName className = ClassName.get(PACKAGE_PREFIX + type.getRef().getModule(), type.getRef().getName());
            if (params.isEmpty()) {
                return className;
            }
            List<TypeName> javaTypes = params.stream().map(s -> s.hasUnit() ? WildcardTypeName.subtypeOf(Object.class) : toJavaTypeName(s, typeAliasMap)).toList();
            return ParameterizedTypeName.get(className, javaTypes.toArray(new TypeName[javaTypes.size()]));
        } else if (type.hasMap()) {
            return ParameterizedTypeName.get(ClassName.get(Map.class), toJavaTypeName(type.getMap().getKey(), typeAliasMap), toJavaTypeName(type.getMap().getValue(), typeAliasMap));
        } else if (type.hasTime()) {
            return ClassName.get(Instant.class);
        } else if (type.hasInt()) {
            return TypeName.LONG;
        } else if (type.hasUnit()) {
            return TypeName.VOID;
        } else if (type.hasBool()) {
            return TypeName.BOOLEAN;
        } else if (type.hasFloat()) {
            return TypeName.DOUBLE;
        } else if (type.hasBytes()) {
            return ArrayTypeName.of(TypeName.BYTE);
        } else if (type.hasAny()) {
            return TypeName.OBJECT;
        }

        throw new RuntimeException("Cannot generate Java type name: " + type);
    }

    @Override
    public boolean shouldRun(Path sourceDir, Config config) {
        return true;
    }

    record Key(String module, String name) {
    }


    static String className(String in) {
        return Character.toUpperCase(in.charAt(0)) + in.substring(1);
    }

    private static final Set<String> JAVA_KEYWORDS = Set.of("abstract", "continue", "for", "new", "switch", "assert",
            "default", "goto", "package", "synchronized", "boolean", "do", "if", "private", "this", "break", "double",
            "implements", "protected", "throw", "byte", "else", "import", "public", "throws", "case", "enum", "instanceof",
            "return", "transient", "catch", "extends", "int", "short", "try", "char", "final", "interface", "static", "void",
            "class", "finally", "long", "strictfp", "volatile", "const", "float", "native", "super", "while");
}
