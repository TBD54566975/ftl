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
import java.util.List;
import java.util.Map;
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

        try (Stream<Path> pathStream = Files.list(context.inputDir())) {
            for (var file : pathStream.toList()) {
                String fileName = file.getFileName().toString();
                if (!fileName.endsWith(".pb")) {
                    continue;
                }
                var module = Module.parseFrom(Files.readAllBytes(file));
                for (var decl : module.getDeclsList()) {
                    if (decl.hasVerb()) {
                        var verb = decl.getVerb();
                        if (!verb.getExport()) {
                            continue;
                        }
                        try {

                            String packageName = PACKAGE_PREFIX + module.getName();
                            TypeSpec.Builder typeBuilder = TypeSpec.interfaceBuilder(className(verb.getName()) + CLIENT)
                                    .addAnnotation(AnnotationSpec.builder(VerbClientDefinition.class)
                                            .addMember("name", "\"" + verb.getName() + "\"")
                                            .addMember("module", "\"" + module.getName() + "\"")
                                            .build())
                                    .addModifiers(Modifier.PUBLIC);
                            if (verb.getRequest().hasUnit() && verb.getResponse().hasUnit()) {
                                typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClientEmpty.class)));
                            } else if (verb.getRequest().hasUnit()) {
                                typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClientSource.class), toJavaTypeName(verb.getResponse())));
                                typeBuilder.addMethod(MethodSpec.methodBuilder("call").returns(toAnnotatedJavaTypeName(verb.getResponse())).addModifiers(Modifier.ABSTRACT).addModifiers(Modifier.PUBLIC).build());
                            } else if (verb.getResponse().hasUnit()) {
                                typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClientSink.class), toJavaTypeName(verb.getRequest())));
                                typeBuilder.addMethod(MethodSpec.methodBuilder("call").returns(TypeName.VOID).addParameter(toAnnotatedJavaTypeName(verb.getRequest()), "value").addModifiers(Modifier.ABSTRACT).addModifiers(Modifier.PUBLIC).build());
                            } else {
                                typeBuilder.addSuperinterface(ParameterizedTypeName.get(ClassName.get(VerbClient.class), toJavaTypeName(verb.getRequest()), toJavaTypeName(verb.getResponse())));
                                typeBuilder.addMethod(MethodSpec.methodBuilder("call").returns(toAnnotatedJavaTypeName(verb.getResponse())).addParameter(toAnnotatedJavaTypeName(verb.getRequest()), "value").addModifiers(Modifier.ABSTRACT).addModifiers(Modifier.PUBLIC).build());
                            }

                            TypeSpec helloWorld = typeBuilder
                                    .build();

                            JavaFile javaFile = JavaFile.builder(packageName, helloWorld)
                                    .build();

                            javaFile.writeTo(context.outDir());

                        } catch (IOException e) {
                            throw new RuntimeException(e);
                        }
                    } else if (decl.hasData()) {
                        var data = decl.getData();
                        try {
                            String packageName = PACKAGE_PREFIX + module.getName();
                            String thisType = className(data.getName());
                            TypeSpec.Builder dataBuilder = TypeSpec.classBuilder(thisType)
                                    .addModifiers(Modifier.PUBLIC);
                            for (var param : data.getTypeParametersList()) {
                                dataBuilder.addTypeVariable(TypeVariableName.get(param.getName()));
                            }

                            for (var i : data.getFieldsList()) {
                                TypeName dataType = toAnnotatedJavaTypeName(i.getType());
                                dataBuilder.addField(dataType, i.getName(), Modifier.PRIVATE);
                                String methodName = Character.toUpperCase(i.getName().charAt(0)) + i.getName().substring(1);
                                dataBuilder.addMethod(MethodSpec.methodBuilder("set" + methodName)
                                                .addModifiers(Modifier.PUBLIC)
                                                .addParameter(dataType, i.getName())
                                                .returns(ClassName.get(packageName, thisType))
                                                .addCode("this.$L = $L;\n", i.getName(), i.getName())
                                                .addCode("return this;")
                                        .build());
                                if (i.getType().hasBool()) {
                                    dataBuilder.addMethod(MethodSpec.methodBuilder("is" + methodName)
                                            .addModifiers(Modifier.PUBLIC)
                                            .returns(dataType)
                                            .addCode("return $L;", i.getName())
                                            .build());
                                } else {
                                    dataBuilder.addMethod(MethodSpec.methodBuilder("get" + methodName)
                                            .addModifiers(Modifier.PUBLIC)
                                            .returns(dataType)
                                            .addCode("return $L;", i.getName())
                                            .build());
                                }
                            }

                            JavaFile javaFile = JavaFile.builder(packageName, dataBuilder.build())
                                    .build();

                            javaFile.writeTo(context.outDir());

                        } catch (IOException e) {
                            throw new RuntimeException(e);
                        }
                    }
                }
            }

        } catch (Exception e) {
            throw new CodeGenException(e);
        }
        return true;
    }

    private TypeName toAnnotatedJavaTypeName(Type type) {
        var results = toJavaTypeName(type);
        if (type.hasRef() || type.hasArray() || type.hasBytes() || type.hasString() || type.hasMap() || type.hasTime()) {
            return results.annotated(AnnotationSpec.builder(NotNull.class).build());
        }
        return results;
    }

    private TypeName toJavaTypeName(Type type) {
        if (type.hasArray()) {
            return ParameterizedTypeName.get(ClassName.get(List.class), toJavaTypeName(type.getArray().getElement()));
        } else if (type.hasString()) {
            return ClassName.get(String.class);
        } else if (type.hasOptional()) {
            return  toJavaTypeName(type.getOptional().getType());
        } else if (type.hasRef()) {
            if (type.getRef().getModule().isEmpty()) {
                return TypeVariableName.get(type.getRef().getName());
            }
            return ClassName.get(PACKAGE_PREFIX + type.getRef().getModule(), type.getRef().getName());
        } else if (type.hasMap()) {
            return ParameterizedTypeName.get(ClassName.get(Map.class), toJavaTypeName(type.getMap().getKey()), toJavaTypeName(type.getMap().getValue()));
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
            return ArrayTypeName.BYTE;
        }

        throw new RuntimeException("Cannot generate Java type name: " + type);
    }

    @Override
    public boolean shouldRun(Path sourceDir, Config config) {
        return true;
    }


    static String className(String in) {
        return Character.toUpperCase(in.charAt(0)) + in.substring(1);
    }
}
