package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;

import org.eclipse.microprofile.config.Config;

import io.quarkus.bootstrap.prebuild.CodeGenException;
import io.quarkus.deployment.CodeGenContext;
import io.quarkus.deployment.CodeGenProvider;
import xyz.block.ftl.v1.schema.Data;
import xyz.block.ftl.v1.schema.Enum;
import xyz.block.ftl.v1.schema.Module;
import xyz.block.ftl.v1.schema.Topic;
import xyz.block.ftl.v1.schema.Type;
import xyz.block.ftl.v1.schema.Verb;

public abstract class JVMCodeGenerator implements CodeGenProvider {

    public static final String PACKAGE_PREFIX = "ftl.";

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
        Map<DeclRef, Type> typeAliasMap = new HashMap<>();
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
                        typeAliasMap.put(new DeclRef(module.getName(), data.getName()), data.getType());
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
                        generateVerb(module, verb, packageName, typeAliasMap, context.outDir());
                    } else if (decl.hasData()) {
                        var data = decl.getData();
                        if (!data.getExport()) {
                            continue;
                        }
                        generateDataObject(module, data, packageName, typeAliasMap, context.outDir());

                    } else if (decl.hasEnum()) {
                        var data = decl.getEnum();
                        if (!data.getExport()) {
                            continue;
                        }
                        generateEnum(module, data, packageName, typeAliasMap, context.outDir());
                    } else if (decl.hasTopic()) {
                        var data = decl.getTopic();
                        if (!data.getExport()) {
                            continue;
                        }
                        generateTopicSubscription(module, data, packageName, typeAliasMap, context.outDir());
                    }
                }
            }

        } catch (Exception e) {
            throw new CodeGenException(e);
        }
        return true;
    }

    protected abstract void generateTopicSubscription(Module module, Topic data, String packageName,
            Map<DeclRef, Type> typeAliasMap, Path outputDir) throws IOException;

    protected abstract void generateEnum(Module module, Enum data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Path outputDir) throws IOException;

    protected abstract void generateDataObject(Module module, Data data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Path outputDir) throws IOException;

    protected abstract void generateVerb(Module module, Verb verb, String packageName, Map<DeclRef, Type> typeAliasMap,
            Path outputDir) throws IOException;

    @Override
    public boolean shouldRun(Path sourceDir, Config config) {
        return true;
    }

    public record DeclRef(String module, String name) {
    }

    protected static String className(String in) {
        return Character.toUpperCase(in.charAt(0)) + in.substring(1);
    }

}
