package xyz.block.ftl.deployment;

import static xyz.block.ftl.deployment.FTLDotNames.ENUM;
import static xyz.block.ftl.deployment.FTLDotNames.EXPORT;
import static xyz.block.ftl.deployment.FTLDotNames.GENERATED_REF;

import java.io.IOException;
import java.io.OutputStream;
import java.lang.reflect.Modifier;
import java.time.Instant;
import java.time.OffsetDateTime;
import java.time.ZonedDateTime;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.function.BiFunction;
import java.util.function.Consumer;
import java.util.regex.Pattern;

import org.jboss.jandex.AnnotationTarget;
import org.jboss.jandex.ArrayType;
import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.ClassType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.IndexView;
import org.jboss.jandex.MethodInfo;
import org.jboss.jandex.PrimitiveType;
import org.jboss.jandex.VoidType;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.processor.DotNames;
import xyz.block.ftl.Config;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.Secret;
import xyz.block.ftl.VerbName;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.builtin.HttpRequest;
import xyz.block.ftl.runtime.builtin.HttpResponse;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.schema.AliasKind;
import xyz.block.ftl.v1.schema.Any;
import xyz.block.ftl.v1.schema.Array;
import xyz.block.ftl.v1.schema.Bool;
import xyz.block.ftl.v1.schema.Bytes;
import xyz.block.ftl.v1.schema.Data;
import xyz.block.ftl.v1.schema.Decl;
import xyz.block.ftl.v1.schema.Field;
import xyz.block.ftl.v1.schema.Float;
import xyz.block.ftl.v1.schema.Int;
import xyz.block.ftl.v1.schema.Metadata;
import xyz.block.ftl.v1.schema.MetadataAlias;
import xyz.block.ftl.v1.schema.MetadataCalls;
import xyz.block.ftl.v1.schema.MetadataConfig;
import xyz.block.ftl.v1.schema.MetadataSecrets;
import xyz.block.ftl.v1.schema.MetadataTypeMap;
import xyz.block.ftl.v1.schema.Module;
import xyz.block.ftl.v1.schema.Position;
import xyz.block.ftl.v1.schema.Ref;
import xyz.block.ftl.v1.schema.Time;
import xyz.block.ftl.v1.schema.Type;
import xyz.block.ftl.v1.schema.TypeAlias;
import xyz.block.ftl.v1.schema.Unit;
import xyz.block.ftl.v1.schema.Verb;

public class ModuleBuilder {

    public static final String BUILTIN = "builtin";

    public static final DotName INSTANT = DotName.createSimple(Instant.class);
    public static final DotName ZONED_DATE_TIME = DotName.createSimple(ZonedDateTime.class);
    public static final DotName NOT_NULL = DotName.createSimple(NotNull.class);
    public static final DotName NULLABLE = DotName.createSimple(Nullable.class);
    public static final DotName JSON_NODE = DotName.createSimple(JsonNode.class.getName());
    public static final DotName OFFSET_DATE_TIME = DotName.createSimple(OffsetDateTime.class.getName());

    private static final Pattern NAME_PATTERN = Pattern.compile("^[A-Za-z_][A-Za-z0-9_]*$");

    private final IndexView index;
    private final Module.Builder protoModuleBuilder;
    private final Map<String, Decl> decls = new HashMap<>();
    private final Map<String, Ref> externalRefs = new HashMap<>();
    private final String moduleName;
    private final Set<String> knownSecrets = new HashSet<>();
    private final Set<String> knownConfig = new HashSet<>();
    private final Map<DotName, TopicsBuildItem.DiscoveredTopic> knownTopics;
    private final Map<DotName, VerbClientBuildItem.DiscoveredClients> verbClients;
    private final FTLRecorder recorder;
    private final CommentsBuildItem comments;
    private final List<ValidationFailure> validationFailures = new ArrayList<>();
    private final boolean defaultToOptional;

    public ModuleBuilder(IndexView index, String moduleName, Map<DotName, TopicsBuildItem.DiscoveredTopic> knownTopics,
            Map<DotName, VerbClientBuildItem.DiscoveredClients> verbClients, FTLRecorder recorder,
            CommentsBuildItem comments, boolean defaultToOptional) {
        this.index = index;
        this.moduleName = moduleName;
        this.protoModuleBuilder = Module.newBuilder()
                .setName(moduleName)
                .setBuiltin(false);
        this.knownTopics = knownTopics;
        this.verbClients = verbClients;
        this.recorder = recorder;
        this.comments = comments;
        this.defaultToOptional = defaultToOptional;
    }

    public static @NotNull String methodToName(MethodInfo method) {
        if (method.hasAnnotation(VerbName.class)) {
            return method.annotation(VerbName.class).value().asString();
        }
        return method.name();
    }

    public String getModuleName() {
        return moduleName;
    }

    public static Class<?> loadClass(org.jboss.jandex.Type param) throws ClassNotFoundException {
        if (param.kind() == org.jboss.jandex.Type.Kind.PARAMETERIZED_TYPE) {
            return Class.forName(param.asParameterizedType().name().toString(), false,
                    Thread.currentThread().getContextClassLoader());
        } else if (param.kind() == org.jboss.jandex.Type.Kind.CLASS) {
            return Class.forName(param.name().toString(), false, Thread.currentThread().getContextClassLoader());
        } else if (param.kind() == org.jboss.jandex.Type.Kind.PRIMITIVE) {
            switch (param.asPrimitiveType().primitive()) {
                case BOOLEAN:
                    return Boolean.TYPE;
                case BYTE:
                    return Byte.TYPE;
                case SHORT:
                    return Short.TYPE;
                case INT:
                    return Integer.TYPE;
                case LONG:
                    return Long.TYPE;
                case FLOAT:
                    return java.lang.Float.TYPE;
                case DOUBLE:
                    return Double.TYPE;
                case CHAR:
                    return Character.TYPE;
                default:
                    throw new RuntimeException("Unknown primitive type " + param.asPrimitiveType().primitive());
            }
        } else if (param.kind() == org.jboss.jandex.Type.Kind.ARRAY) {
            ArrayType array = param.asArrayType();
            if (array.componentType().kind() == org.jboss.jandex.Type.Kind.PRIMITIVE) {
                switch (array.componentType().asPrimitiveType().primitive()) {
                    case BOOLEAN:
                        return boolean[].class;
                    case BYTE:
                        return byte[].class;
                    case SHORT:
                        return short[].class;
                    case INT:
                        return int[].class;
                    case LONG:
                        return long[].class;
                    case FLOAT:
                        return float[].class;
                    case DOUBLE:
                        return double[].class;
                    case CHAR:
                        return char[].class;
                    default:
                        throw new RuntimeException("Unknown primitive type " + param.asPrimitiveType().primitive());
                }
            }
        }
        throw new RuntimeException("Unknown type " + param.kind());

    }

    public void registerVerbMethod(MethodInfo method, String className,
            boolean exported, BodyType bodyType, Consumer<Verb.Builder> metadataCallback) {
        try {
            List<Class<?>> parameterTypes = new ArrayList<>();
            List<BiFunction<ObjectMapper, CallRequest, Object>> paramMappers = new ArrayList<>();
            org.jboss.jandex.Type bodyParamType = null;
            Nullability bodyParamNullability = Nullability.MISSING;

            xyz.block.ftl.v1.schema.Verb.Builder verbBuilder = xyz.block.ftl.v1.schema.Verb.newBuilder();
            String verbName = validateName(className, ModuleBuilder.methodToName(method));
            MetadataCalls.Builder callsMetadata = MetadataCalls.newBuilder();
            MetadataConfig.Builder configMetadata = MetadataConfig.newBuilder();
            MetadataSecrets.Builder secretMetadata = MetadataSecrets.newBuilder();
            for (var param : method.parameters()) {
                if (param.hasAnnotation(Secret.class)) {
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    String name = param.annotation(Secret.class).value().asString();
                    paramMappers.add(new VerbRegistry.SecretSupplier(name, paramType));
                    if (!knownSecrets.contains(name)) {
                        xyz.block.ftl.v1.schema.Secret.Builder secretBuilder = xyz.block.ftl.v1.schema.Secret.newBuilder()
                                .setType(buildType(param.type(), false, param))
                                .setName(name)
                                .addAllComments(comments.getComments(name));
                        addDecls(Decl.newBuilder().setSecret(secretBuilder).build());
                        knownSecrets.add(name);
                    }
                    secretMetadata.addSecrets(Ref.newBuilder().setName(name).setModule(moduleName).build());
                } else if (param.hasAnnotation(Config.class)) {
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    String name = param.annotation(Config.class).value().asString();
                    paramMappers.add(new VerbRegistry.ConfigSupplier(name, paramType));
                    if (!knownConfig.contains(name)) {
                        xyz.block.ftl.v1.schema.Config.Builder configBuilder = xyz.block.ftl.v1.schema.Config.newBuilder()
                                .setType(buildType(param.type(), false, param))
                                .setName(name)
                                .addAllComments(comments.getComments(name));
                        addDecls(Decl.newBuilder().setConfig(configBuilder).build());
                        knownConfig.add(name);
                    }
                    configMetadata.addConfig(Ref.newBuilder().setName(name).setModule(moduleName).build());
                } else if (knownTopics.containsKey(param.type().name())) {
                    var topic = knownTopics.get(param.type().name());
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    paramMappers.add(recorder.topicSupplier(topic.generatedProducer(), verbName));
                } else if (verbClients.containsKey(param.type().name())) {
                    var client = verbClients.get(param.type().name());
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    paramMappers.add(recorder.verbClientSupplier(client.generatedClient()));
                    callsMetadata.addCalls(Ref.newBuilder().setName(client.name()).setModule(client.module()).build());
                } else if (FTLDotNames.LEASE_CLIENT.equals(param.type().name())) {
                    parameterTypes.add(LeaseClient.class);
                    paramMappers.add(recorder.leaseClientSupplier());
                } else if (bodyType != BodyType.DISALLOWED && bodyParamType == null) {
                    bodyParamType = param.type();
                    bodyParamNullability = nullability(param);
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    //TODO: map and list types
                    paramMappers.add(new VerbRegistry.BodySupplier(paramType));
                } else {
                    throw new RuntimeException("Unknown parameter type " + param.type() + " on FTL method: "
                            + method.declaringClass().name() + "." + method.name());
                }
            }
            if (bodyParamType == null) {
                if (bodyType == BodyType.REQUIRED) {
                    throw new RuntimeException("Missing required payload parameter");
                }
                bodyParamType = VoidType.VOID;
            }
            if (callsMetadata.getCallsCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setCalls(callsMetadata));
            }
            if (secretMetadata.getSecretsCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setSecrets(secretMetadata));
            }
            if (configMetadata.getConfigCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setConfig(configMetadata));
            }

            recorder.registerVerb(moduleName, verbName, method.name(), parameterTypes,
                    Class.forName(className, false, Thread.currentThread().getContextClassLoader()), paramMappers,
                    method.returnType() == VoidType.VOID);

            verbBuilder.setName(verbName)
                    .setExport(exported)
                    .setPos(PositionUtils.forMethod(method))
                    .setRequest(buildType(bodyParamType, exported, bodyParamNullability))
                    .setResponse(buildType(method.returnType(), exported, method))
                    .addAllComments(comments.getComments(verbName));
            if (metadataCallback != null) {
                metadataCallback.accept(verbBuilder);
            }
            addDecls(Decl.newBuilder().setVerb(verbBuilder)
                    .build());

        } catch (Exception e) {
            throw new RuntimeException("Failed to process FTL method " + method.declaringClass().name() + "." + method.name(),
                    e);
        }
    }

    private Nullability nullability(org.jboss.jandex.AnnotationTarget type) {
        if (type.hasAnnotation(NULLABLE)) {
            return Nullability.NULLABLE;
        } else if (type.hasAnnotation(NOT_NULL)) {
            return Nullability.NOT_NULL;
        }
        return Nullability.MISSING;
    }

    private Type handleNullabilityAnnotations(Type res, Nullability nullability) {
        if (nullability == Nullability.NOT_NULL) {
            return res;
        } else if (nullability == Nullability.NULLABLE || defaultToOptional) {
            return Type.newBuilder().setOptional(xyz.block.ftl.v1.schema.Optional.newBuilder()
                    .setType(res))
                    .build();
        }
        return res;
    }

    public Type buildType(org.jboss.jandex.Type type, boolean export, AnnotationTarget target) {
        return buildType(type, export, nullability(target));
    }

    public Type buildType(org.jboss.jandex.Type type, boolean export, Nullability nullability) {
        switch (type.kind()) {
            case PRIMITIVE -> {
                var prim = type.asPrimitiveType();
                switch (prim.primitive()) {
                    case INT, LONG, BYTE, SHORT -> {
                        return Type.newBuilder().setInt(Int.newBuilder().build()).build();
                    }
                    case FLOAT, DOUBLE -> {
                        return Type.newBuilder().setFloat(Float.newBuilder().build()).build();
                    }
                    case BOOLEAN -> {
                        return Type.newBuilder().setBool(Bool.newBuilder().build()).build();
                    }
                    case CHAR -> {
                        return Type.newBuilder().setString(xyz.block.ftl.v1.schema.String.newBuilder().build()).build();
                    }
                    default -> throw new RuntimeException("unknown primitive type: " + prim.primitive());
                }
            }
            case VOID -> {
                return Type.newBuilder().setUnit(Unit.newBuilder().build()).build();
            }
            case ARRAY -> {
                ArrayType arrayType = type.asArrayType();
                if (arrayType.componentType().kind() == org.jboss.jandex.Type.Kind.PRIMITIVE && arrayType
                        .componentType().asPrimitiveType().primitive() == PrimitiveType.Primitive.BYTE) {
                    return handleNullabilityAnnotations(Type.newBuilder().setBytes(Bytes.newBuilder().build()).build(),
                            nullability);
                }
                return handleNullabilityAnnotations(Type.newBuilder()
                        .setArray(Array.newBuilder()
                                .setElement(buildType(arrayType.componentType(), export, Nullability.NOT_NULL)).build())
                        .build(), nullability);
            }
            case CLASS -> {
                var clazz = type.asClassType();
                var info = index.getClassByName(clazz.name());
                if (info.enclosingClass() != null && !Modifier.isStatic(info.flags())) {
                    // proceed as normal, we fail at the end
                    validationFailures.add(new ValidationFailure(clazz.name().toString(),
                            "Inner classes must be static"));
                }

                PrimitiveType unboxed = PrimitiveType.unbox(clazz);
                if (unboxed != null) {
                    Type primitive = buildType(unboxed, export, Nullability.NOT_NULL);
                    if (nullability == Nullability.NOT_NULL) {
                        return primitive;
                    }
                    return Type.newBuilder().setOptional(xyz.block.ftl.v1.schema.Optional.newBuilder()
                            .setType(primitive))
                            .build();
                }
                if (info.hasDeclaredAnnotation(GENERATED_REF)) {
                    var ref = info.declaredAnnotation(GENERATED_REF);
                    return handleNullabilityAnnotations(Type.newBuilder()
                            .setRef(Ref.newBuilder().setName(ref.value("name").asString())
                                    .setModule(ref.value("module").asString()))
                            .build(), nullability);
                }
                if (clazz.name().equals(DotName.STRING_NAME)) {
                    return handleNullabilityAnnotations(
                            Type.newBuilder().setString(xyz.block.ftl.v1.schema.String.newBuilder().build()).build(),
                            nullability);
                }
                if (clazz.name().equals(DotName.OBJECT_NAME) || clazz.name().equals(JSON_NODE)) {
                    return handleNullabilityAnnotations(Type.newBuilder().setAny(Any.newBuilder().build()).build(),
                            nullability);
                }
                if (clazz.name().equals(OFFSET_DATE_TIME) || clazz.name().equals(INSTANT)
                        || clazz.name().equals(ZONED_DATE_TIME)) {
                    return handleNullabilityAnnotations(Type.newBuilder().setTime(Time.newBuilder().build()).build(),
                            nullability);
                }

                String name = clazz.name().local();
                if (externalRefs.containsKey(name)) {
                    // Ref is to another module. Don't need a Decl
                    return Type.newBuilder().setRef(externalRefs.get(name)).build();
                }
                var ref = Type.newBuilder().setRef(
                        Ref.newBuilder().setName(name).setModule(moduleName).build()).build();

                if (info.isEnum() || info.hasAnnotation(ENUM)) {
                    // Set only the name and export here. EnumProcessor will fill in the rest
                    xyz.block.ftl.v1.schema.Enum.Builder ennum = xyz.block.ftl.v1.schema.Enum.newBuilder()
                            .setName(name)
                            .setExport(type.hasAnnotation(EXPORT) || export);
                    addDecls(Decl.newBuilder().setEnum(ennum.build()).build());
                    return ref;
                } else {
                    // If this data was processed already, skip early
                    if (setDeclExport(name, type.hasAnnotation(EXPORT) || export)) {
                        return ref;
                    }
                    Data.Builder data = Data.newBuilder()
                            .setPos(PositionUtils.forClass(clazz.name().toString()))
                            .setName(name)
                            .setExport(type.hasAnnotation(EXPORT) || export)
                            .addAllComments(comments.getComments(name));
                    buildDataElement(data, clazz.name());
                    addDecls(Decl.newBuilder().setData(data).build());
                    return ref;
                }
            }
            case PARAMETERIZED_TYPE -> {
                var paramType = type.asParameterizedType();
                if (paramType.name().equals(DotName.createSimple(List.class))) {
                    return handleNullabilityAnnotations(Type.newBuilder()
                            .setArray(Array.newBuilder()
                                    .setElement(buildType(paramType.arguments().get(0), export, Nullability.NOT_NULL)))
                            .build(), nullability);
                } else if (paramType.name().equals(DotName.createSimple(Map.class))) {
                    return handleNullabilityAnnotations(Type.newBuilder().setMap(xyz.block.ftl.v1.schema.Map.newBuilder()
                            .setKey(buildType(paramType.arguments().get(0), export, Nullability.NOT_NULL))
                            .setValue(buildType(paramType.arguments().get(1), export, Nullability.NOT_NULL)))
                            .build(), nullability);
                } else if (paramType.name().equals(DotNames.OPTIONAL)) {
                    //TODO: optional kinda sucks
                    return Type.newBuilder().setOptional(xyz.block.ftl.v1.schema.Optional.newBuilder()
                            .setType(buildType(paramType.arguments().get(0), export, Nullability.NOT_NULL)))
                            .build();
                } else if (paramType.name().equals(DotName.createSimple(HttpRequest.class))) {
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpRequest.class.getSimpleName())
                                    .addTypeParameters(buildType(paramType.arguments().get(0), export, Nullability.NOT_NULL)))
                            .build();
                } else if (paramType.name().equals(DotName.createSimple(HttpResponse.class))) {
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpResponse.class.getSimpleName())
                                    .addTypeParameters(buildType(paramType.arguments().get(0), export, Nullability.NOT_NULL))
                                    .addTypeParameters(Type.newBuilder().setUnit(Unit.newBuilder().build())))
                            .build();
                } else {
                    ClassInfo classByName = index.getClassByName(paramType.name());
                    validateName(classByName.name().toString(), classByName.name().local());
                    var cb = ClassType.builder(classByName.name());
                    var main = buildType(cb.build(), export, Nullability.NOT_NULL);
                    var builder = main.toBuilder();
                    var refBuilder = builder.getRef().toBuilder();

                    for (var arg : paramType.arguments()) {
                        refBuilder.addTypeParameters(buildType(arg, export, Nullability.NOT_NULL));
                    }

                    builder.setRef(refBuilder);
                    return handleNullabilityAnnotations(builder.build(), nullability);
                }
            }
        }

        throw new RuntimeException("NOT YET IMPLEMENTED");
    }

    private void buildDataElement(Data.Builder data, DotName className) {
        if (className == null || className.equals(DotName.OBJECT_NAME)) {
            return;
        }
        var clazz = index.getClassByName(className);
        if (clazz == null) {
            return;
        }
        //TODO: handle getters and setters properly, also Jackson annotations etc
        for (var field : clazz.fields()) {
            if (!Modifier.isStatic(field.flags())) {
                Field.Builder builder = Field.newBuilder().setName(field.name())
                        .setType(buildType(field.type(), data.getExport(), field));
                if (field.hasAnnotation(JsonAlias.class)) {
                    var aliases = field.annotation(JsonAlias.class);
                    if (aliases.value() != null) {
                        for (var alias : aliases.value().asStringArray()) {
                            builder.addMetadata(
                                    Metadata.newBuilder().setAlias(
                                            MetadataAlias.newBuilder().setKind(AliasKind.ALIAS_KIND_JSON).setAlias(alias)));
                        }
                    }
                }
                data.addFields(builder.build());
            }
        }
        buildDataElement(data, clazz.superName());
    }

    public ModuleBuilder addDecls(Decl decl) {
        if (decl.hasData()) {
            Data data = decl.getData();
            if (!setDeclExport(data.getName(), data.getExport())) {
                addDecl(decl, data.getPos(), data.getName());
            }
        } else if (decl.hasEnum()) {
            xyz.block.ftl.v1.schema.Enum enuum = decl.getEnum();
            if (!updateEnum(enuum.getName(), decl)) {
                addDecl(decl, enuum.getPos(), enuum.getName());
            }
        } else if (decl.hasDatabase()) {
            addDecl(decl, decl.getDatabase().getPos(), decl.getDatabase().getName());
        } else if (decl.hasConfig()) {
            addDecl(decl, decl.getConfig().getPos(), decl.getConfig().getName());
        } else if (decl.hasSecret()) {
            addDecl(decl, decl.getSecret().getPos(), decl.getSecret().getName());
        } else if (decl.hasVerb()) {
            addDecl(decl, decl.getVerb().getPos(), decl.getVerb().getName());
        } else if (decl.hasTypeAlias()) {
            addDecl(decl, decl.getTypeAlias().getPos(), decl.getTypeAlias().getName());
        } else if (decl.hasTopic()) {
            addDecl(decl, decl.getTopic().getPos(), decl.getTopic().getName());
        } else if (decl.hasFsm()) {
            addDecl(decl, decl.getFsm().getPos(), decl.getFsm().getName());
        } else if (decl.hasSubscription()) {
            addDecl(decl, decl.getSubscription().getPos(), decl.getSubscription().getName());
        }

        return this;
    }

    public int getDeclsCount() {
        return decls.size();
    }

    public void writeTo(OutputStream out) throws IOException {
        decls.values().stream().forEachOrdered(protoModuleBuilder::addDecls);

        if (!validationFailures.isEmpty()) {
            StringBuilder sb = new StringBuilder();
            for (var failure : validationFailures) {
                sb.append("Validation failure: ").append(failure.className).append(": ").append(failure.message)
                        .append("\n");
            }
            throw new RuntimeException(sb.toString());
        }
        protoModuleBuilder.build().writeTo(out);
    }

    public void registerTypeAlias(String name, org.jboss.jandex.Type finalT, org.jboss.jandex.Type finalS, boolean exported,
            Map<String, String> languageMappings) {
        validateName(finalT.name().toString(), name);
        TypeAlias.Builder typeAlias = TypeAlias.newBuilder()
                .setType(buildType(finalS, exported, Nullability.NOT_NULL))
                .setName(name)
                .addAllComments(comments.getComments(name))
                .addMetadata(Metadata.newBuilder()
                        .setTypeMap(MetadataTypeMap.newBuilder().setRuntime("java").setNativeName(finalT.toString())
                                .build())
                        .build());
        for (var entry : languageMappings.entrySet()) {
            typeAlias.addMetadata(Metadata.newBuilder().setTypeMap(MetadataTypeMap.newBuilder().setRuntime(entry.getKey())
                    .setNativeName(entry.getValue()).build()).build());
        }
        addDecls(Decl.newBuilder().setTypeAlias(typeAlias).build());
    }

    /**
     * Types from other modules don't need a Decl. We store Ref for it, and prevent a Decl being created next
     * time we see this name
     */
    public void registerExternalType(String module, String name) {
        Ref ref = Ref.newBuilder()
                .setModule(module)
                .setName(name)
                .build();
        externalRefs.put(name, ref);
    }

    private void addDecl(Decl decl, Position pos, String name) {
        validateName(pos, name);
        if (decls.containsKey(name)) {
            duplicateNameValidationError(name, pos);
        }
        decls.put(name, decl);
    }

    /**
     * Check if an enum with the given name already exists in the module. If it does, merge fields from both into one
     */
    private boolean updateEnum(String name, Decl decl) {
        if (decls.containsKey(name)) {
            var existing = decls.get(name);
            if (!existing.hasEnum()) {
                duplicateNameValidationError(name, decl.getEnum().getPos());
            }
            var moreComplete = decl.getEnum().getVariantsCount() > 0 ? decl : existing;
            var lessComplete = decl.getEnum().getVariantsCount() > 0 ? existing : decl;
            boolean export = lessComplete.getEnum().getExport() || existing.getEnum().getExport();
            var merged = moreComplete.getEnum().toBuilder()
                    .setExport(export)
                    .build();
            decls.put(name, Decl.newBuilder().setEnum(merged).build());
            if (export) {
                // Need to update export on variants too
                for (var childDecl : merged.getVariantsList()) {
                    if (childDecl.getValue().hasTypeValue() && childDecl.getValue().getTypeValue().getValue().hasRef()) {
                        var ref = childDecl.getValue().getTypeValue().getValue().getRef();
                        setDeclExport(ref.getName(), true);
                    }
                }
            }
            return true;
        }
        return false;
    }

    /**
     * Set a Decl's export field to <code>export</code>. Return true iff the Decl exists
     */
    private boolean setDeclExport(String name, boolean export) {
        var existing = decls.get(name);
        if (existing != null) {
            if (existing.hasData()) {
                var merged = existing.getData().toBuilder().setExport(export).build();
                decls.put(name, Decl.newBuilder().setData(merged).build());
            } else if (existing.hasTypeAlias()) {
                var merged = existing.getTypeAlias().toBuilder().setExport(export).build();
                decls.put(name, Decl.newBuilder().setTypeAlias(merged).build());
            }

        }
        return existing != null;
    }

    private void duplicateNameValidationError(String name, Position pos) {
        validationFailures.add(new ValidationFailure(name, String.format(
                "schema declaration with name \"%s\" already exists for module \"%s\"; previously declared at \"%s\"",
                name, moduleName, pos.getFilename() + ":" + pos.getLine())));
    }

    public enum BodyType {
        DISALLOWED,
        ALLOWED,
        REQUIRED
    }

    record ValidationFailure(String className, String message) {
    }

    String validateName(Position position, String name) {
        //we group all validation failures together so we can report them all at once
        if (!NAME_PATTERN.matcher(name).matches()) {
            validationFailures.add(
                    new ValidationFailure(position == null ? "<unknown>" : position.getFilename() + ":" + position.getLine(),
                            String.format("Invalid name %s, must match " + NAME_PATTERN, name)));
        }
        return name;
    }

    String validateName(String className, String name) {
        //we group all validation failures together so we can report them all at once
        if (!NAME_PATTERN.matcher(name).matches()) {
            validationFailures
                    .add(new ValidationFailure(className, String.format("Invalid name %s, must match " + NAME_PATTERN, name)));
        }
        return name;
    }
}
