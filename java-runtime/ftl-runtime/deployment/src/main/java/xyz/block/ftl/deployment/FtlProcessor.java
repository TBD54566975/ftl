package xyz.block.ftl.deployment;

import java.io.IOException;
import java.lang.reflect.Modifier;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.PosixFilePermission;
import java.time.OffsetDateTime;
import java.util.ArrayList;
import java.util.EnumSet;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.function.BiFunction;
import java.util.function.Consumer;
import java.util.stream.Collectors;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.AnnotationTarget;
import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.ClassType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.MethodInfo;
import org.jboss.jandex.VoidType;
import org.jboss.logging.Logger;
import org.jboss.resteasy.reactive.common.model.MethodParameter;
import org.jboss.resteasy.reactive.common.model.ParameterType;
import org.jboss.resteasy.reactive.server.core.parameters.ParameterExtractor;
import org.jboss.resteasy.reactive.server.mapping.URITemplate;
import org.jboss.resteasy.reactive.server.processor.scanning.MethodScanner;
import org.jetbrains.annotations.NotNull;

import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.arc.processor.DotNames;
import io.quarkus.deployment.Capabilities;
import io.quarkus.deployment.Capability;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.ApplicationInfoBuildItem;
import io.quarkus.deployment.builditem.ApplicationStartBuildItem;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.FeatureBuildItem;
import io.quarkus.deployment.builditem.LaunchModeBuildItem;
import io.quarkus.deployment.builditem.ShutdownContextBuildItem;
import io.quarkus.deployment.builditem.SystemPropertyBuildItem;
import io.quarkus.deployment.builditem.nativeimage.ReflectiveClassBuildItem;
import io.quarkus.deployment.pkg.builditem.OutputTargetBuildItem;
import io.quarkus.grpc.deployment.BindableServiceBuildItem;
import io.quarkus.netty.runtime.virtual.VirtualServerChannel;
import io.quarkus.resteasy.reactive.server.deployment.ResteasyReactiveResourceMethodEntriesBuildItem;
import io.quarkus.resteasy.reactive.server.spi.MethodScannerBuildItem;
import io.quarkus.vertx.core.deployment.CoreVertxBuildItem;
import io.quarkus.vertx.core.deployment.EventLoopCountBuildItem;
import io.quarkus.vertx.http.deployment.RequireVirtualHttpBuildItem;
import io.quarkus.vertx.http.deployment.WebsocketSubProtocolsBuildItem;
import io.quarkus.vertx.http.runtime.HttpBuildTimeConfig;
import io.quarkus.vertx.http.runtime.VertxHttpRecorder;
import xyz.block.ftl.Config;
import xyz.block.ftl.Cron;
import xyz.block.ftl.Export;
import xyz.block.ftl.GeneratedRef;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.Retry;
import xyz.block.ftl.Secret;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.Verb;
import xyz.block.ftl.VerbName;
import xyz.block.ftl.runtime.FTLController;
import xyz.block.ftl.runtime.FTLHttpHandler;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.JsonSerializationConfig;
import xyz.block.ftl.runtime.TopicHelper;
import xyz.block.ftl.runtime.VerbClientHelper;
import xyz.block.ftl.runtime.VerbHandler;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.builtin.HttpRequest;
import xyz.block.ftl.runtime.builtin.HttpResponse;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.schema.Array;
import xyz.block.ftl.v1.schema.Bool;
import xyz.block.ftl.v1.schema.Data;
import xyz.block.ftl.v1.schema.Decl;
import xyz.block.ftl.v1.schema.Field;
import xyz.block.ftl.v1.schema.Float;
import xyz.block.ftl.v1.schema.IngressPathComponent;
import xyz.block.ftl.v1.schema.IngressPathLiteral;
import xyz.block.ftl.v1.schema.IngressPathParameter;
import xyz.block.ftl.v1.schema.Int;
import xyz.block.ftl.v1.schema.Metadata;
import xyz.block.ftl.v1.schema.MetadataCalls;
import xyz.block.ftl.v1.schema.MetadataCronJob;
import xyz.block.ftl.v1.schema.MetadataIngress;
import xyz.block.ftl.v1.schema.MetadataRetry;
import xyz.block.ftl.v1.schema.MetadataSubscriber;
import xyz.block.ftl.v1.schema.Module;
import xyz.block.ftl.v1.schema.Optional;
import xyz.block.ftl.v1.schema.Ref;
import xyz.block.ftl.v1.schema.Time;
import xyz.block.ftl.v1.schema.Type;
import xyz.block.ftl.v1.schema.Unit;

class FtlProcessor {

    private static final Logger log = Logger.getLogger(FtlProcessor.class);

    private static final String SCHEMA_OUT = "schema.pb";
    private static final String FEATURE = "ftl-java-runtime";
    public static final DotName EXPORT = DotName.createSimple(Export.class);
    public static final DotName VERB = DotName.createSimple(Verb.class);
    public static final DotName CRON = DotName.createSimple(Cron.class);
    public static final DotName SUBSCRIPTION = DotName.createSimple(Subscription.class);
    public static final String BUILTIN = "builtin";
    public static final DotName CONSUMER = DotName.createSimple(Consumer.class);
    public static final DotName SECRET = DotName.createSimple(Secret.class);
    public static final DotName CONFIG = DotName.createSimple(Config.class);
    public static final DotName OFFSET_DATE_TIME = DotName.createSimple(OffsetDateTime.class.getName());
    public static final DotName GENERATED_REF = DotName.createSimple(GeneratedRef.class);
    public static final DotName LEASE_CLIENT = DotName.createSimple(LeaseClient.class);

    @BuildStep
    ModuleNameBuildItem moduleName(ApplicationInfoBuildItem applicationInfoBuildItem) {
        return new ModuleNameBuildItem(applicationInfoBuildItem.getName());

    }

    @BuildStep
    FeatureBuildItem feature() {
        return new FeatureBuildItem(FEATURE);
    }

    @BuildStep
    BindableServiceBuildItem verbService() {
        var ret = new BindableServiceBuildItem(DotName.createSimple(VerbHandler.class));
        ret.registerBlockingMethod("call");
        ret.registerBlockingMethod("publishEvent");
        ret.registerBlockingMethod("sendFSMEvent");
        ret.registerBlockingMethod("acquireLease");
        ret.registerBlockingMethod("getModuleContext");
        ret.registerBlockingMethod("ping");
        return ret;
    }

    @BuildStep
    AdditionalBeanBuildItem beans() {
        return AdditionalBeanBuildItem.builder()
                .addBeanClasses(VerbHandler.class,
                        VerbRegistry.class, FTLHttpHandler.class, FTLController.class,
                        TopicHelper.class, VerbClientHelper.class, JsonSerializationConfig.class)
                .setUnremovable().build();
    }

    @BuildStep
    AdditionalBeanBuildItem verbBeans(CombinedIndexBuildItem index) {

        var beans = AdditionalBeanBuildItem.builder().setUnremovable();
        for (var verb : index.getIndex().getAnnotations(VERB)) {
            beans.addBeanClasses(verb.target().asMethod().declaringClass().name().toString());
        }
        return beans.build();
    }

    @BuildStep
    public SystemPropertyBuildItem moduleNameConfig(ApplicationInfoBuildItem applicationInfoBuildItem) {
        return new SystemPropertyBuildItem("ftl.module.name", applicationInfoBuildItem.getName());
    }

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    public MethodScannerBuildItem methodScanners(TopicsBuildItem topics,
            VerbClientBuildItem verbClients, FTLRecorder recorder) {
        return new MethodScannerBuildItem(new MethodScanner() {
            @Override
            public ParameterExtractor handleCustomParameter(org.jboss.jandex.Type type,
                    Map<DotName, AnnotationInstance> annotations, boolean field, Map<String, Object> methodContext) {
                try {

                    if (annotations.containsKey(SECRET)) {
                        Class<?> paramType = loadClass(type);
                        String name = annotations.get(SECRET).value().asString();
                        return new VerbRegistry.SecretSupplier(name, paramType);
                    } else if (annotations.containsKey(CONFIG)) {
                        Class<?> paramType = loadClass(type);
                        String name = annotations.get(CONFIG).value().asString();
                        return new VerbRegistry.ConfigSupplier(name, paramType);
                    } else if (topics.getTopics().containsKey(type.name())) {
                        var topic = topics.getTopics().get(type.name());
                        return recorder.topicParamExtractor(topic.generatedProducer());
                    } else if (verbClients.getVerbClients().containsKey(type.name())) {
                        var client = verbClients.getVerbClients().get(type.name());
                        return recorder.verbParamExtractor(client.generatedClient());
                    } else if (LEASE_CLIENT.equals(type.name())) {
                        return recorder.leaseClientExtractor();
                    }
                    return null;
                } catch (ClassNotFoundException e) {
                    throw new RuntimeException(e);
                }
            }
        });
    }

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    public void registerVerbs(CombinedIndexBuildItem index,
            FTLRecorder recorder,
            OutputTargetBuildItem outputTargetBuildItem,
            ResteasyReactiveResourceMethodEntriesBuildItem restEndpoints,
            TopicsBuildItem topics,
            VerbClientBuildItem verbClients,
            ModuleNameBuildItem moduleNameBuildItem,
            SubscriptionMetaAnnotationsBuildItem subscriptionMetaAnnotationsBuildItem) throws Exception {
        String moduleName = moduleNameBuildItem.getModuleName();
        Module.Builder moduleBuilder = Module.newBuilder()
                .setName(moduleName)
                .setBuiltin(false);
        Map<TypeKey, Ref> dataElements = new HashMap<>();
        ExtractionContext extractionContext = new ExtractionContext(moduleName, index, recorder, moduleBuilder, dataElements,
                new HashSet<>(), new HashSet<>(), topics.getTopics(), verbClients.getVerbClients());
        var beans = AdditionalBeanBuildItem.builder().setUnremovable();

        //register all the topics we are defining in the module definition

        for (var topic : topics.getTopics().values()) {
            extractionContext.moduleBuilder.addDecls(Decl.newBuilder().setTopic(xyz.block.ftl.v1.schema.Topic.newBuilder()
                    .setExport(topic.exported())
                    .setName(topic.topicName())
                    .setEvent(buildType(extractionContext, topic.eventType())).build()));
        }

        handleVerbAnnotations(index, beans, extractionContext);
        handleCronAnnotations(index, beans, extractionContext);
        handleSubscriptionAnnotations(index, subscriptionMetaAnnotationsBuildItem, moduleName, moduleBuilder, extractionContext,
                beans);

        //TODO: make this composable so it is not just one big method, build items should contribute to the schema
        for (var endpoint : restEndpoints.getEntries()) {
            //TODO: naming
            var verbName = methodToName(endpoint.getMethodInfo());
            recorder.registerHttpIngress(moduleName, verbName);

            //TODO: handle type parameters properly
            org.jboss.jandex.Type bodyParamType = VoidType.VOID;
            MethodParameter[] parameters = endpoint.getResourceMethod().getParameters();
            for (int i = 0, parametersLength = parameters.length; i < parametersLength; i++) {
                var param = parameters[i];
                if (param.parameterType.equals(ParameterType.BODY)) {
                    bodyParamType = endpoint.getMethodInfo().parameterType(i);
                    break;
                }
            }

            StringBuilder pathBuilder = new StringBuilder();
            if (endpoint.getBasicResourceClassInfo().getPath() != null) {
                pathBuilder.append(endpoint.getBasicResourceClassInfo().getPath());
            }
            if (endpoint.getResourceMethod().getPath() != null && !endpoint.getResourceMethod().getPath().isEmpty()) {
                if (pathBuilder.charAt(pathBuilder.length() - 1) != '/'
                        && !endpoint.getResourceMethod().getPath().startsWith("/")) {
                    pathBuilder.append('/');
                }
                pathBuilder.append(endpoint.getResourceMethod().getPath());
            }
            String path = pathBuilder.toString();
            URITemplate template = new URITemplate(path, false);
            List<IngressPathComponent> pathComponents = new ArrayList<>();
            for (var i : template.components) {
                if (i.type == URITemplate.Type.CUSTOM_REGEX) {
                    throw new RuntimeException(
                            "Invalid path " + path + " on HTTP endpoint: " + endpoint.getActualClassInfo().name() + "."
                                    + methodToName(endpoint.getMethodInfo())
                                    + " FTL does not support custom regular expressions");
                } else if (i.type == URITemplate.Type.LITERAL) {
                    if (i.literalText.equals("/")) {
                        continue;
                    }
                    pathComponents.add(IngressPathComponent.newBuilder()
                            .setIngressPathLiteral(IngressPathLiteral.newBuilder().setText(i.literalText.replace("/", "")))
                            .build());
                } else {
                    pathComponents.add(IngressPathComponent.newBuilder()
                            .setIngressPathParameter(IngressPathParameter.newBuilder().setName(i.name.replace("/", "")))
                            .build());
                }
            }

            //TODO: process path properly
            MetadataIngress.Builder ingressBuilder = MetadataIngress.newBuilder()
                    .setMethod(endpoint.getResourceMethod().getHttpMethod());
            for (var i : pathComponents) {
                ingressBuilder.addPath(i);
            }
            Metadata ingressMetadata = Metadata.newBuilder()
                    .setIngress(ingressBuilder
                            .build())
                    .build();
            Type requestTypeParam = buildType(extractionContext, bodyParamType);
            Type responseTypeParam = buildType(extractionContext, endpoint.getMethodInfo().returnType());
            moduleBuilder
                    .addDecls(Decl.newBuilder().setVerb(xyz.block.ftl.v1.schema.Verb.newBuilder()
                            .addMetadata(ingressMetadata)
                            .setName(verbName)
                            .setExport(true)
                            .setRequest(Type.newBuilder()
                                    .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpRequest.class.getSimpleName())
                                            .addTypeParameters(requestTypeParam))
                                    .build())
                            .setResponse(Type.newBuilder()
                                    .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpResponse.class.getSimpleName())
                                            .addTypeParameters(responseTypeParam)
                                            .addTypeParameters(Type.newBuilder().setUnit(Unit.newBuilder())))
                                    .build()))
                            .build());
        }

        Path output = outputTargetBuildItem.getOutputDirectory().resolve(SCHEMA_OUT);
        try (var out = Files.newOutputStream(output)) {
            moduleBuilder.build().writeTo(out);
        }

        output = outputTargetBuildItem.getOutputDirectory().resolve("main");
        try (var out = Files.newOutputStream(output)) {
            out.write("""
                    #!/bin/bash
                    exec java -jar quarkus-app/quarkus-run.jar""".getBytes(StandardCharsets.UTF_8));
        }
        var perms = Files.getPosixFilePermissions(output);
        EnumSet<PosixFilePermission> newPerms = EnumSet.copyOf(perms);
        newPerms.add(PosixFilePermission.GROUP_EXECUTE);
        newPerms.add(PosixFilePermission.OWNER_EXECUTE);
        Files.setPosixFilePermissions(output, newPerms);
    }

    private void handleVerbAnnotations(CombinedIndexBuildItem index, AdditionalBeanBuildItem.Builder beans,
            ExtractionContext extractionContext) {
        for (var verb : index.getIndex().getAnnotations(VERB)) {
            boolean exported = verb.target().hasAnnotation(EXPORT);
            var method = verb.target().asMethod();
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);

            handleVerbMethod(extractionContext, method, className, exported, BodyType.ALLOWED, null);
        }
    }

    private void handleSubscriptionAnnotations(CombinedIndexBuildItem index,
            SubscriptionMetaAnnotationsBuildItem subscriptionMetaAnnotationsBuildItem, String moduleName,
            Module.Builder moduleBuilder, ExtractionContext extractionContext, AdditionalBeanBuildItem.Builder beans) {
        for (var subscription : index.getIndex().getAnnotations(SUBSCRIPTION)) {
            var info = SubscriptionMetaAnnotationsBuildItem.fromJandex(subscription, moduleName);
            if (subscription.target().kind() != AnnotationTarget.Kind.METHOD) {
                continue;
            }
            var method = subscription.target().asMethod();
            String className = method.declaringClass().name().toString();
            generateSubscription(moduleBuilder, extractionContext, beans, method, className, info);
        }
        for (var metaSub : subscriptionMetaAnnotationsBuildItem.getAnnotations().entrySet()) {
            for (var subscription : index.getIndex().getAnnotations(metaSub.getKey())) {
                if (subscription.target().kind() != AnnotationTarget.Kind.METHOD) {
                    log.warnf("Subscription annotation on non-method target: %s", subscription.target());
                    continue;
                }
                var method = subscription.target().asMethod();
                generateSubscription(moduleBuilder, extractionContext, beans, method,
                        method.declaringClass().name().toString(),
                        metaSub.getValue());
            }

        }
    }

    private void handleCronAnnotations(CombinedIndexBuildItem index, AdditionalBeanBuildItem.Builder beans,
            ExtractionContext extractionContext) {
        for (var cron : index.getIndex().getAnnotations(CRON)) {
            var method = cron.target().asMethod();
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);
            handleVerbMethod(extractionContext, method, className, false, BodyType.DISALLOWED, (builder -> {
                builder.addMetadata(Metadata.newBuilder()
                        .setCronJob(MetadataCronJob.newBuilder().setCron(cron.value().asString())).build());
            }));
        }
    }

    private void generateSubscription(Module.Builder moduleBuilder, ExtractionContext extractionContext,
            AdditionalBeanBuildItem.Builder beans, MethodInfo method, String className,
            SubscriptionMetaAnnotationsBuildItem.SubscriptionAnnotation info) {
        beans.addBeanClass(className);
        moduleBuilder.addDecls(Decl.newBuilder().setSubscription(xyz.block.ftl.v1.schema.Subscription.newBuilder()
                .setName(info.name()).setTopic(Ref.newBuilder().setName(info.topic()).setModule(info.module()).build()))
                .build());
        handleVerbMethod(extractionContext, method, className, false, BodyType.REQUIRED, (builder -> {
            builder.addMetadata(Metadata.newBuilder().setSubscriber(MetadataSubscriber.newBuilder().setName(info.name())));
            if (method.hasAnnotation(Retry.class)) {
                RetryRecord retry = RetryRecord.fromJandex(method.annotation(Retry.class), extractionContext.moduleName);

                MetadataRetry.Builder retryBuilder = MetadataRetry.newBuilder();
                if (!retry.catchVerb().isEmpty()) {
                    retryBuilder.setCatch(Ref.newBuilder().setModule(retry.catchModule())
                            .setName(retry.catchVerb()).build());
                }
                retryBuilder.setCount(retry.count())
                        .setMaxBackoff(retry.maxBackoff())
                        .setMinBackoff(retry.minBackoff());
                builder.addMetadata(Metadata.newBuilder().setRetry(retryBuilder).build());
            }
        }));
    }

    private void handleVerbMethod(ExtractionContext context, MethodInfo method, String className,
            boolean exported, BodyType bodyType, Consumer<xyz.block.ftl.v1.schema.Verb.Builder> metadataCallback) {
        try {
            List<Class<?>> parameterTypes = new ArrayList<>();
            List<BiFunction<ObjectMapper, CallRequest, Object>> paramMappers = new ArrayList<>();
            org.jboss.jandex.Type bodyParamType = null;
            xyz.block.ftl.v1.schema.Verb.Builder verbBuilder = xyz.block.ftl.v1.schema.Verb.newBuilder();
            String verbName = methodToName(method);
            MetadataCalls.Builder callsMetadata = MetadataCalls.newBuilder();
            for (var param : method.parameters()) {
                if (param.hasAnnotation(Secret.class)) {
                    Class<?> paramType = loadClass(param.type());
                    parameterTypes.add(paramType);
                    String name = param.annotation(Secret.class).value().asString();
                    paramMappers.add(new VerbRegistry.SecretSupplier(name, paramType));
                    if (!context.knownSecrets.contains(name)) {
                        context.moduleBuilder.addDecls(Decl.newBuilder().setSecret(xyz.block.ftl.v1.schema.Secret.newBuilder()
                                .setType(buildType(context, param.type())).setName(name)));
                        context.knownSecrets.add(name);
                    }
                } else if (param.hasAnnotation(Config.class)) {
                    Class<?> paramType = loadClass(param.type());
                    parameterTypes.add(paramType);
                    String name = param.annotation(Config.class).value().asString();
                    paramMappers.add(new VerbRegistry.ConfigSupplier(name, paramType));
                    if (!context.knownConfig.contains(name)) {
                        context.moduleBuilder.addDecls(Decl.newBuilder().setConfig(xyz.block.ftl.v1.schema.Config.newBuilder()
                                .setType(buildType(context, param.type())).setName(name)));
                        context.knownConfig.add(name);
                    }
                } else if (context.knownTopics.containsKey(param.type().name())) {
                    var topic = context.knownTopics.get(param.type().name());
                    Class<?> paramType = loadClass(param.type());
                    parameterTypes.add(paramType);
                    paramMappers.add(context.recorder().topicSupplier(topic.generatedProducer(), verbName));
                } else if (context.verbClients.containsKey(param.type().name())) {
                    var client = context.verbClients.get(param.type().name());
                    Class<?> paramType = loadClass(param.type());
                    parameterTypes.add(paramType);
                    paramMappers.add(context.recorder().verbClientSupplier(client.generatedClient()));
                    callsMetadata.addCalls(Ref.newBuilder().setName(client.name()).setModule(client.module()).build());
                } else if (LEASE_CLIENT.equals(param.type().name())) {
                    parameterTypes.add(LeaseClient.class);
                    paramMappers.add(context.recorder().leaseClientSupplier());
                } else if (bodyType != BodyType.DISALLOWED && bodyParamType == null) {
                    bodyParamType = param.type();
                    Class<?> paramType = loadClass(param.type());
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

            //TODO: we need better handling around Optional
            context.recorder.registerVerb(context.moduleName(), verbName, method.name(), parameterTypes,
                    Class.forName(className, false, Thread.currentThread().getContextClassLoader()), paramMappers,
                    method.returnType() == VoidType.VOID);
            verbBuilder
                    .setName(verbName)
                    .setExport(exported)
                    .setRequest(buildType(context, bodyParamType))
                    .setResponse(buildType(context, method.returnType()));

            if (metadataCallback != null) {
                metadataCallback.accept(verbBuilder);
            }
            context.moduleBuilder
                    .addDecls(Decl.newBuilder().setVerb(verbBuilder)
                            .build());

        } catch (Exception e) {
            throw new RuntimeException("Failed to process FTL method " + method.declaringClass().name() + "." + method.name(),
                    e);
        }
    }

    private static @NotNull String methodToName(MethodInfo method) {
        if (method.hasAnnotation(VerbName.class)) {
            return method.annotation(VerbName.class).value().asString();
        }
        return method.name();
    }

    private static Class<?> loadClass(org.jboss.jandex.Type param) throws ClassNotFoundException {
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
                    return java.lang.Double.TYPE;
                case CHAR:
                    return Character.TYPE;
                default:
                    throw new RuntimeException("Unknown primitive type " + param.asPrimitiveType().primitive());
            }
        } else {
            throw new RuntimeException("Unknown type " + param.kind());
        }
    }

    /**
     * This is a huge hack that is needed until Quarkus supports both virtual and socket based HTTP
     */
    @Record(ExecutionTime.RUNTIME_INIT)
    @BuildStep
    void openSocket(ApplicationStartBuildItem start,
            LaunchModeBuildItem launchMode,
            CoreVertxBuildItem vertx,
            ShutdownContextBuildItem shutdown,
            BuildProducer<ReflectiveClassBuildItem> reflectiveClass,
            HttpBuildTimeConfig httpBuildTimeConfig,
            java.util.Optional<RequireVirtualHttpBuildItem> requireVirtual,
            EventLoopCountBuildItem eventLoopCount,
            List<WebsocketSubProtocolsBuildItem> websocketSubProtocols,
            Capabilities capabilities,
            VertxHttpRecorder recorder) throws IOException {
        reflectiveClass
                .produce(ReflectiveClassBuildItem.builder(VirtualServerChannel.class)
                        .build());
        recorder.startServer(vertx.getVertx(), shutdown,
                launchMode.getLaunchMode(), true, false,
                eventLoopCount.getEventLoopCount(),
                websocketSubProtocols.stream().map(bi -> bi.getWebsocketSubProtocols())
                        .collect(Collectors.toList()),
                launchMode.isAuxiliaryApplication(), !capabilities.isPresent(Capability.VERTX_WEBSOCKETS));
    }

    private Type buildType(ExtractionContext context, org.jboss.jandex.Type type) {
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
                return Type.newBuilder()
                        .setArray(Array.newBuilder().setElement(buildType(context, type.asArrayType().componentType())).build())
                        .build();
            }
            case CLASS -> {
                var clazz = type.asClassType();
                var info = context.index().getComputingIndex().getClassByName(clazz.name());
                if (info != null && info.hasDeclaredAnnotation(GENERATED_REF)) {
                    var ref = info.declaredAnnotation(GENERATED_REF);
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setName(ref.value("name").asString())
                                    .setModule(ref.value("module").asString()))
                            .build();
                }
                if (clazz.name().equals(DotName.STRING_NAME)) {
                    return Type.newBuilder().setString(xyz.block.ftl.v1.schema.String.newBuilder().build()).build();
                }
                if (clazz.name().equals(OFFSET_DATE_TIME)) {
                    return Type.newBuilder().setTime(Time.newBuilder().build()).build();
                }
                var existing = context.dataElements.get(new TypeKey(clazz.name().toString(), List.of()));
                if (existing != null) {
                    return Type.newBuilder().setRef(existing).build();
                }
                Data.Builder data = Data.newBuilder();
                data.setName(clazz.name().local());
                data.setExport(type.hasAnnotation(EXPORT));
                buildDataElement(context, data, clazz.name());
                context.moduleBuilder.addDecls(Decl.newBuilder().setData(data).build());
                Ref ref = Ref.newBuilder().setName(data.getName()).setModule(context.moduleName).build();
                context.dataElements.put(new TypeKey(clazz.name().toString(), List.of()), ref);
                return Type.newBuilder().setRef(ref).build();
            }
            case PARAMETERIZED_TYPE -> {
                var paramType = type.asParameterizedType();
                if (paramType.name().equals(DotName.createSimple(List.class))) {
                    return Type.newBuilder()
                            .setArray(Array.newBuilder().setElement(buildType(context, paramType.arguments().get(0)))).build();
                } else if (paramType.name().equals(DotName.createSimple(Map.class))) {
                    return Type.newBuilder().setMap(xyz.block.ftl.v1.schema.Map.newBuilder()
                            .setKey(buildType(context, paramType.arguments().get(0)))
                            .setValue(buildType(context, paramType.arguments().get(0))))
                            .build();
                } else if (paramType.name().equals(DotNames.OPTIONAL)) {
                    return Type.newBuilder()
                            .setOptional(Optional.newBuilder().setType(buildType(context, paramType.arguments().get(0))))
                            .build();
                } else if (paramType.name().equals(DotName.createSimple(HttpRequest.class))) {
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpRequest.class.getSimpleName())
                                    .addTypeParameters(buildType(context, paramType.arguments().get(0))))
                            .build();
                } else if (paramType.name().equals(DotName.createSimple(HttpResponse.class))) {
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpResponse.class.getSimpleName())
                                    .addTypeParameters(buildType(context, paramType.arguments().get(0)))
                                    .addTypeParameters(Type.newBuilder().setUnit(Unit.newBuilder().build())))
                            .build();
                } else {
                    ClassInfo classByName = context.index().getComputingIndex().getClassByName(paramType.name());
                    var cb = ClassType.builder(classByName.name());
                    var main = buildType(context, cb.build());
                    var builder = main.toBuilder();
                    var refBuilder = builder.getRef().toBuilder();

                    for (var arg : paramType.arguments()) {
                        refBuilder.addTypeParameters(buildType(context, arg));
                    }
                    builder.setRef(refBuilder);
                    return builder.build();
                }
            }
        }

        throw new RuntimeException("NOT YET IMPLEMENTED");
    }

    private void buildDataElement(ExtractionContext context, Data.Builder data, DotName className) {
        if (className == null || className.equals(DotName.OBJECT_NAME)) {
            return;
        }
        var clazz = context.index.getComputingIndex().getClassByName(className);
        if (clazz == null) {
            return;
        }
        //TODO: handle getters and setters properly, also Jackson annotations etc
        for (var field : clazz.fields()) {
            if (!Modifier.isStatic(field.flags())) {
                data.addFields(Field.newBuilder().setName(field.name()).setType(buildType(context, field.type())).build());
            }
        }
        buildDataElement(context, data, clazz.superName());
    }

    private record TypeKey(String name, List<String> typeParams) {

    }

    record ExtractionContext(String moduleName, CombinedIndexBuildItem index, FTLRecorder recorder,
            Module.Builder moduleBuilder,
            Map<TypeKey, Ref> dataElements, Set<String> knownSecrets, Set<String> knownConfig,
            Map<DotName, TopicsBuildItem.DiscoveredTopic> knownTopics,
            Map<DotName, VerbClientBuildItem.DiscoveredClients> verbClients) {
    }

    enum BodyType {
        DISALLOWED,
        ALLOWED,
        REQUIRED
    }
}
