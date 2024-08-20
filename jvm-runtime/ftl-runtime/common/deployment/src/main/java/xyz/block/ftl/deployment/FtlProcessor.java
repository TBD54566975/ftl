package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.PosixFilePermission;
import java.util.ArrayList;
import java.util.EnumSet;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.function.BiFunction;
import java.util.function.Consumer;
import java.util.stream.Collectors;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.AnnotationTarget;
import org.jboss.jandex.ArrayType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.MethodInfo;
import org.jboss.jandex.PrimitiveType;
import org.jboss.jandex.VoidType;
import org.jboss.logging.Logger;
import org.jboss.resteasy.reactive.common.model.MethodParameter;
import org.jboss.resteasy.reactive.common.model.ParameterType;
import org.jboss.resteasy.reactive.server.core.parameters.ParameterExtractor;
import org.jboss.resteasy.reactive.server.mapping.URITemplate;
import org.jboss.resteasy.reactive.server.processor.scanning.MethodScanner;
import org.jetbrains.annotations.NotNull;

import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.agroal.spi.JdbcDataSourceBuildItem;
import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.deployment.Capabilities;
import io.quarkus.deployment.Capability;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.ApplicationInfoBuildItem;
import io.quarkus.deployment.builditem.ApplicationStartBuildItem;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.GeneratedResourceBuildItem;
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
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.Retry;
import xyz.block.ftl.Secret;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.Verb;
import xyz.block.ftl.VerbName;
import xyz.block.ftl.runtime.FTLDatasourceCredentials;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.JsonSerializationConfig;
import xyz.block.ftl.runtime.TopicHelper;
import xyz.block.ftl.runtime.VerbClientHelper;
import xyz.block.ftl.runtime.VerbHandler;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.builtin.HttpRequest;
import xyz.block.ftl.runtime.builtin.HttpResponse;
import xyz.block.ftl.runtime.config.FTLConfigSource;
import xyz.block.ftl.runtime.http.FTLHttpHandler;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.schema.Array;
import xyz.block.ftl.v1.schema.Database;
import xyz.block.ftl.v1.schema.Decl;
import xyz.block.ftl.v1.schema.Float;
import xyz.block.ftl.v1.schema.IngressPathComponent;
import xyz.block.ftl.v1.schema.IngressPathLiteral;
import xyz.block.ftl.v1.schema.IngressPathParameter;
import xyz.block.ftl.v1.schema.Metadata;
import xyz.block.ftl.v1.schema.MetadataCalls;
import xyz.block.ftl.v1.schema.MetadataCronJob;
import xyz.block.ftl.v1.schema.MetadataIngress;
import xyz.block.ftl.v1.schema.MetadataRetry;
import xyz.block.ftl.v1.schema.MetadataSubscriber;
import xyz.block.ftl.v1.schema.Optional;
import xyz.block.ftl.v1.schema.Ref;
import xyz.block.ftl.v1.schema.Type;
import xyz.block.ftl.v1.schema.Unit;

class FtlProcessor {

    private static final Logger log = Logger.getLogger(FtlProcessor.class);

    private static final String SCHEMA_OUT = "schema.pb";

    public static final String BUILTIN = "builtin";
    public static final DotName EXPORT = DotName.createSimple(Export.class);
    public static final DotName VERB = DotName.createSimple(Verb.class);
    public static final DotName CRON = DotName.createSimple(Cron.class);
    public static final DotName SUBSCRIPTION = DotName.createSimple(Subscription.class);
    public static final DotName SECRET = DotName.createSimple(Secret.class);
    public static final DotName CONFIG = DotName.createSimple(Config.class);
    public static final DotName LEASE_CLIENT = DotName.createSimple(LeaseClient.class);

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
                        VerbRegistry.class, FTLHttpHandler.class,
                        TopicHelper.class, VerbClientHelper.class, JsonSerializationConfig.class,
                        FTLDatasourceCredentials.class)
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
            SubscriptionMetaAnnotationsBuildItem subscriptionMetaAnnotationsBuildItem,
            List<JdbcDataSourceBuildItem> datasources,
            BuildProducer<SystemPropertyBuildItem> systemPropProducer,
            BuildProducer<GeneratedResourceBuildItem> generatedResourceBuildItemBuildProducer) throws Exception {
        String moduleName = moduleNameBuildItem.getModuleName();

        SchemaBuilder schemaBuilder = new SchemaBuilder(index.getComputingIndex(), moduleName);

        ExtractionContext extractionContext = new ExtractionContext(moduleName, index, recorder, schemaBuilder,
                new HashSet<>(), new HashSet<>(), topics.getTopics(), verbClients.getVerbClients());
        var beans = AdditionalBeanBuildItem.builder().setUnremovable();

        List<String> namedDatasources = new ArrayList<>();
        for (var ds : datasources) {
            if (!ds.getDbKind().equals("postgresql")) {
                throw new RuntimeException("only postgresql is supported not " + ds.getDbKind());
            }
            //default name is <default> which is not a valid name
            String sanitisedName = ds.getName().replace("<", "").replace(">", "");
            //we use a dynamic credentials provider
            if (ds.isDefault()) {
                systemPropProducer
                        .produce(new SystemPropertyBuildItem("quarkus.datasource.credentials-provider", sanitisedName));
                systemPropProducer
                        .produce(new SystemPropertyBuildItem("quarkus.datasource.credentials-provider-name",
                                FTLDatasourceCredentials.NAME));
            } else {
                namedDatasources.add(ds.getName());
                systemPropProducer.produce(new SystemPropertyBuildItem(
                        "quarkus.datasource." + ds.getName() + ".credentials-provider", sanitisedName));
                systemPropProducer.produce(new SystemPropertyBuildItem(
                        "quarkus.datasource." + ds.getName() + ".credentials-provider-name", FTLDatasourceCredentials.NAME));
            }
            schemaBuilder.addDecls(
                    Decl.newBuilder().setDatabase(
                            Database.newBuilder().setType("postgres").setName(sanitisedName))
                            .build());
        }
        generatedResourceBuildItemBuildProducer.produce(new GeneratedResourceBuildItem(FTLConfigSource.DATASOURCE_NAMES,
                String.join("\n", namedDatasources).getBytes(StandardCharsets.UTF_8)));

        //register all the topics we are defining in the module definition
        for (var topic : topics.getTopics().values()) {
            extractionContext.schemaBuilder.addDecls(Decl.newBuilder().setTopic(xyz.block.ftl.v1.schema.Topic.newBuilder()
                    .setExport(topic.exported())
                    .setName(topic.topicName())
                    .setEvent(schemaBuilder.buildType(topic.eventType(), topic.exported())).build()).build());
        }

        handleVerbAnnotations(index, beans, extractionContext);
        handleCronAnnotations(index, beans, extractionContext);
        handleSubscriptionAnnotations(index, subscriptionMetaAnnotationsBuildItem, moduleName, extractionContext,
                beans);

        //TODO: make this composable so it is not just one big method, build items should contribute to the schema
        for (var endpoint : restEndpoints.getEntries()) {
            //TODO: naming
            var verbName = methodToName(endpoint.getMethodInfo());
            boolean base64 = false;

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

            if (bodyParamType instanceof ArrayType) {
                org.jboss.jandex.Type component = ((ArrayType) bodyParamType).component();
                if (component instanceof PrimitiveType) {
                    base64 = component.asPrimitiveType().equals(PrimitiveType.BYTE);
                }
            }

            recorder.registerHttpIngress(moduleName, verbName, base64);

            StringBuilder pathBuilder = new StringBuilder();
            if (endpoint.getBasicResourceClassInfo().getPath() != null) {
                pathBuilder.append(endpoint.getBasicResourceClassInfo().getPath());
            }
            if (endpoint.getResourceMethod().getPath() != null && !endpoint.getResourceMethod().getPath().isEmpty()) {
                boolean builderEndsSlash = pathBuilder.charAt(pathBuilder.length() - 1) == '/';
                boolean pathStartsSlash = endpoint.getResourceMethod().getPath().startsWith("/");
                if (builderEndsSlash && pathStartsSlash) {
                    pathBuilder.setLength(pathBuilder.length() - 1);
                } else if (!builderEndsSlash && !pathStartsSlash) {
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
                    for (var part : i.literalText.split("/")) {
                        if (part.isEmpty()) {
                            continue;
                        }
                        pathComponents.add(IngressPathComponent.newBuilder()
                                .setIngressPathLiteral(IngressPathLiteral.newBuilder().setText(part))
                                .build());
                    }
                } else {
                    pathComponents.add(IngressPathComponent.newBuilder()
                            .setIngressPathParameter(IngressPathParameter.newBuilder().setName(i.name))
                            .build());
                }
            }

            //TODO: process path properly
            MetadataIngress.Builder ingressBuilder = MetadataIngress.newBuilder()
                    .setType("http")
                    .setMethod(endpoint.getResourceMethod().getHttpMethod());
            for (var i : pathComponents) {
                ingressBuilder.addPath(i);
            }
            Metadata ingressMetadata = Metadata.newBuilder()
                    .setIngress(ingressBuilder
                            .build())
                    .build();
            Type requestTypeParam = schemaBuilder.buildType(bodyParamType, true);
            Type responseTypeParam = schemaBuilder.buildType(endpoint.getMethodInfo().returnType(), true);
            Type stringType = Type.newBuilder().setString(xyz.block.ftl.v1.schema.String.newBuilder().build()).build();
            Type pathParamType = Type.newBuilder()
                    .setMap(xyz.block.ftl.v1.schema.Map.newBuilder().setKey(stringType)
                            .setValue(stringType))
                    .build();
            schemaBuilder
                    .addDecls(Decl.newBuilder().setVerb(xyz.block.ftl.v1.schema.Verb.newBuilder()
                            .addMetadata(ingressMetadata)
                            .setName(verbName)
                            .setExport(true)
                            .setRequest(Type.newBuilder()
                                    .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpRequest.class.getSimpleName())
                                            .addTypeParameters(requestTypeParam)
                                            .addTypeParameters(pathParamType)
                                            .addTypeParameters(Type.newBuilder()
                                                    .setMap(xyz.block.ftl.v1.schema.Map.newBuilder().setKey(stringType)
                                                            .setValue(Type.newBuilder()
                                                                    .setArray(Array.newBuilder().setElement(stringType)))
                                                            .build())))
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
            schemaBuilder.writeTo(out);
        }

        output = outputTargetBuildItem.getOutputDirectory().resolve("main");
        try (var out = Files.newOutputStream(output)) {
            out.write(
                    """
                            #!/bin/bash
                            exec java $FTL_JVM_OPTS -jar quarkus-app/quarkus-run.jar"""
                            .getBytes(StandardCharsets.UTF_8));
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
            ExtractionContext extractionContext, AdditionalBeanBuildItem.Builder beans) {
        for (var subscription : index.getIndex().getAnnotations(SUBSCRIPTION)) {
            var info = SubscriptionMetaAnnotationsBuildItem.fromJandex(subscription, moduleName);
            if (subscription.target().kind() != AnnotationTarget.Kind.METHOD) {
                continue;
            }
            var method = subscription.target().asMethod();
            String className = method.declaringClass().name().toString();
            generateSubscription(extractionContext, beans, method, className, info);
        }
        for (var metaSub : subscriptionMetaAnnotationsBuildItem.getAnnotations().entrySet()) {
            for (var subscription : index.getIndex().getAnnotations(metaSub.getKey())) {
                if (subscription.target().kind() != AnnotationTarget.Kind.METHOD) {
                    log.warnf("Subscription annotation on non-method target: %s", subscription.target());
                    continue;
                }
                var method = subscription.target().asMethod();
                generateSubscription(extractionContext, beans, method,
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

    private void generateSubscription(ExtractionContext extractionContext,
            AdditionalBeanBuildItem.Builder beans, MethodInfo method, String className,
            SubscriptionMetaAnnotationsBuildItem.SubscriptionAnnotation info) {
        beans.addBeanClass(className);
        extractionContext.schemaBuilder.addDecls(Decl.newBuilder()
                .setSubscription(xyz.block.ftl.v1.schema.Subscription.newBuilder()
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
                        context.schemaBuilder.addDecls(Decl.newBuilder().setSecret(xyz.block.ftl.v1.schema.Secret.newBuilder()
                                .setType(context.schemaBuilder.buildType(param.type(), false)).setName(name)).build());
                        context.knownSecrets.add(name);
                    }
                } else if (param.hasAnnotation(Config.class)) {
                    Class<?> paramType = loadClass(param.type());
                    parameterTypes.add(paramType);
                    String name = param.annotation(Config.class).value().asString();
                    paramMappers.add(new VerbRegistry.ConfigSupplier(name, paramType));
                    if (!context.knownConfig.contains(name)) {
                        context.schemaBuilder.addDecls(Decl.newBuilder().setConfig(xyz.block.ftl.v1.schema.Config.newBuilder()
                                .setType(context.schemaBuilder.buildType(param.type(), false)).setName(name)).build());
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
                    .setRequest(context.schemaBuilder.buildType(bodyParamType, exported))
                    .setResponse(context.schemaBuilder.buildType(method.returnType(), exported));

            if (metadataCallback != null) {
                metadataCallback.accept(verbBuilder);
            }
            context.schemaBuilder
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

    record ExtractionContext(String moduleName, CombinedIndexBuildItem index, FTLRecorder recorder,
            SchemaBuilder schemaBuilder,
            Set<String> knownSecrets, Set<String> knownConfig,
            Map<DotName, TopicsBuildItem.DiscoveredTopic> knownTopics,
            Map<DotName, VerbClientBuildItem.DiscoveredClients> verbClients) {
    }

    enum BodyType {
        DISALLOWED,
        ALLOWED,
        REQUIRED
    }

}
