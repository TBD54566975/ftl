package xyz.block.ftl.runtime;

import java.net.URI;
import java.net.URISyntaxException;
import java.time.Duration;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.regex.Pattern;

import org.jboss.logging.Logger;

import com.google.protobuf.ByteString;

import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.v1.*;
import xyz.block.ftl.v1.schema.Ref;

public class FTLController implements LeaseClient {
    private static final Logger log = Logger.getLogger(FTLController.class);
    final String moduleName;

    private Throwable currentError;
    private volatile ModuleContextResponse moduleContextResponse;
    private boolean waiters = false;

    final VerbServiceGrpc.VerbServiceStub verbService;
    final ModuleServiceGrpc.ModuleServiceStub moduleService;
    final StreamObserver<ModuleContextResponse> moduleObserver = new ModuleObserver();

    private static volatile FTLController controller;

    private final Map<String, ModuleContextResponse.DBType> databases = new ConcurrentHashMap<>();

    /**
     * TODO: look at how init should work, this is terrible and will break dev mode
     */
    public static FTLController instance() {
        if (controller == null) {
            synchronized (FTLController.class) {
                if (controller == null) {
                    controller = new FTLController();
                }
            }
        }
        return controller;
    }

    FTLController() {
        String endpoint = System.getenv("FTL_ENDPOINT");
        String testEndpoint = System.getProperty("ftl.test.endpoint"); //set by the test framework
        if (testEndpoint != null) {
            endpoint = testEndpoint;
        }
        if (endpoint == null) {
            endpoint = "http://localhost:8892";
        }
        var uri = URI.create(endpoint);
        this.moduleName = System.getProperty("ftl.module.name");
        var channelBuilder = ManagedChannelBuilder.forAddress(uri.getHost(), uri.getPort());
        if (uri.getScheme().equals("http")) {
            channelBuilder.usePlaintext();
        }
        var channel = channelBuilder.build();
        moduleService = ModuleServiceGrpc.newStub(channel);
        moduleService.getModuleContext(ModuleContextRequest.newBuilder().setModule(moduleName).build(), moduleObserver);
        verbService = VerbServiceGrpc.newStub(channel);
    }

    public void registerDatabase(String name, ModuleContextResponse.DBType type) {
        databases.put(name, type);
    }

    public byte[] getSecret(String secretName) {
        var context = getModuleContext();
        if (context.containsSecrets(secretName)) {
            return context.getSecretsMap().get(secretName).toByteArray();
        }
        throw new RuntimeException("Secret not found: " + secretName);
    }

    public byte[] getConfig(String secretName) {
        var context = getModuleContext();
        if (context.containsConfigs(secretName)) {
            return context.getConfigsMap().get(secretName).toByteArray();
        }
        throw new RuntimeException("Config not found: " + secretName);
    }

    public Datasource getDatasource(String name) {
        if (databases.get(name) == ModuleContextResponse.DBType.POSTGRES) {
            var proxyAddress = System.getenv("FTL_PROXY_POSTGRES_ADDRESS");
            return new Datasource("jdbc:postgresql://" + proxyAddress + "/" + name, "ftl", "ftl");
        }
        List<ModuleContextResponse.DSN> databasesList = getModuleContext().getDatabasesList();
        for (var i : databasesList) {
            if (i.getName().equals(name)) {
                return Datasource.fromDSN(i.getDsn(), i.getType());
            }
        }
        return null;
    }

    public byte[] callVerb(String name, String module, byte[] payload) {
        CompletableFuture<byte[]> cf = new CompletableFuture<>();

        verbService.call(CallRequest.newBuilder().setVerb(Ref.newBuilder().setModule(module).setName(name))
                .setBody(ByteString.copyFrom(payload)).build(), new StreamObserver<>() {

                    @Override
                    public void onNext(CallResponse callResponse) {
                        if (callResponse.hasError()) {
                            cf.completeExceptionally(new RuntimeException(callResponse.getError().getMessage()));
                        } else {
                            cf.complete(callResponse.getBody().toByteArray());
                        }
                    }

                    @Override
                    public void onError(Throwable throwable) {
                        cf.completeExceptionally(throwable);
                    }

                    @Override
                    public void onCompleted() {

                    }
                });
        try {
            return cf.get();
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public void publishEvent(String topic, String callingVerbName, byte[] event) {
        CompletableFuture<?> cf = new CompletableFuture<>();
        moduleService.publishEvent(PublishEventRequest.newBuilder()
                .setCaller(callingVerbName).setBody(ByteString.copyFrom(event))
                .setTopic(Ref.newBuilder().setModule(moduleName).setName(topic).build()).build(),
                new StreamObserver<PublishEventResponse>() {
                    @Override
                    public void onNext(PublishEventResponse publishEventResponse) {
                        cf.complete(null);
                    }

                    @Override
                    public void onError(Throwable throwable) {
                        cf.completeExceptionally(throwable);
                    }

                    @Override
                    public void onCompleted() {
                        cf.complete(null);
                    }
                });
        try {
            cf.get();
        } catch (InterruptedException | ExecutionException e) {
            throw new RuntimeException(e);
        }
    }

    public LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException {
        CompletableFuture<?> cf = new CompletableFuture<>();
        var client = moduleService.acquireLease(new StreamObserver<AcquireLeaseResponse>() {
            @Override
            public void onNext(AcquireLeaseResponse value) {
                cf.complete(null);
            }

            @Override
            public void onError(Throwable t) {
                cf.completeExceptionally(t);
            }

            @Override
            public void onCompleted() {
                if (!cf.isDone()) {
                    onError(new RuntimeException("stream closed"));
                }
            }
        });
        client.onNext(AcquireLeaseRequest.newBuilder().setModule(moduleName)
                .addAllKey(Arrays.asList(keys))
                .setTtl(com.google.protobuf.Duration.newBuilder()
                        .setSeconds(duration.toSeconds()))
                .build());
        try {
            cf.get();
        } catch (Exception e) {
            throw new LeaseFailedException("lease already held", e);
        }
        return new LeaseHandle() {
            @Override
            public void close() {
                client.onCompleted();
            }
        };
    }

    private ModuleContextResponse getModuleContext() {
        var moduleContext = moduleContextResponse;
        if (moduleContext != null) {
            return moduleContext;
        }
        synchronized (moduleObserver) {
            for (;;) {
                moduleContext = moduleContextResponse;
                if (moduleContext != null) {
                    return moduleContext;
                }
                if (currentError != null) {
                    throw new RuntimeException(currentError);
                }
                waiters = true;
                try {
                    moduleObserver.wait();
                } catch (InterruptedException e) {
                    throw new RuntimeException(e);
                }
            }

        }
    }

    private class ModuleObserver implements StreamObserver<ModuleContextResponse> {

        final AtomicInteger failCount = new AtomicInteger();

        @Override
        public void onNext(ModuleContextResponse moduleContextResponse) {
            synchronized (this) {
                currentError = null;
                FTLController.this.moduleContextResponse = moduleContextResponse;
                if (waiters) {
                    this.notifyAll();
                    waiters = false;
                }
            }

        }

        @Override
        public void onError(Throwable throwable) {
            log.error("GRPC connection error", throwable);
            synchronized (this) {
                currentError = throwable;
                if (waiters) {
                    this.notifyAll();
                    waiters = false;
                }
            }
            if (failCount.incrementAndGet() < 5) {
                moduleService.getModuleContext(ModuleContextRequest.newBuilder().setModule(moduleName).build(), moduleObserver);
            }
        }

        @Override
        public void onCompleted() {
            onError(new RuntimeException("connection closed"));
        }
    }

    public record Datasource(String connectionString, String username, String password) {

        public static Datasource fromDSN(String dsn, ModuleContextResponse.DBType type) {
            String prefix = type.equals(ModuleContextResponse.DBType.MYSQL) ? "jdbc:mysql" : "jdbc:postgresql";
            try {
                URI uri = new URI(dsn);
                String username = "";
                String password = "";
                String userInfo = uri.getUserInfo();
                if (userInfo != null) {
                    var split = userInfo.split(":");
                    username = split[0];
                    password = split[1];
                    return new Datasource(
                            new URI(prefix, null, uri.getHost(), uri.getPort(), uri.getPath(), uri.getQuery(), null)
                                    .toASCIIString(),
                            username, password);
                } else {
                    //TODO: this is horrible, just quick hack for now
                    var matcher = Pattern.compile("[&?]user=([^?&]*)").matcher(dsn);
                    if (matcher.find()) {
                        username = matcher.group(1);
                        dsn = matcher.replaceAll("");
                    }
                    matcher = Pattern.compile("[&?]password=([^?&]*)").matcher(dsn);
                    if (matcher.find()) {
                        password = matcher.group(1);
                        dsn = matcher.replaceAll("");
                    }
                    matcher = Pattern.compile("^([^:]+):([^:]+)@").matcher(dsn);
                    if (matcher.find()) {
                        username = matcher.group(1);
                        password = matcher.group(2);
                        dsn = matcher.replaceAll("");
                    }
                    matcher = Pattern.compile("tcp\\(([^:)]+):([^:)]+)\\)").matcher(dsn);
                    if (matcher.find()) {
                        // Mysql has a messed up syntax
                        dsn = matcher.replaceAll(matcher.group(1) + ":" + matcher.group(2));
                    }
                    dsn = dsn.replaceAll("postgresql://", "");
                    dsn = dsn.replaceAll("postgres://", "");
                    dsn = dsn.replaceAll("mysql://", "");
                    dsn = prefix + "://" + dsn;
                    return new Datasource(dsn, username, password);
                }
            } catch (URISyntaxException e) {
                throw new RuntimeException(e);
            }
        }

    }
}
