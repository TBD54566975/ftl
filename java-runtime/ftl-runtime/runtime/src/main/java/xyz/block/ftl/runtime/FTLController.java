package xyz.block.ftl.runtime;

import java.net.URI;
import java.time.Duration;
import java.util.Arrays;
import java.util.Deque;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.LinkedBlockingDeque;

import jakarta.annotation.PreDestroy;
import jakarta.inject.Singleton;

import org.eclipse.microprofile.config.inject.ConfigProperty;
import org.jboss.logging.Logger;

import com.google.protobuf.ByteString;

import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import io.quarkus.runtime.Startup;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.v1.AcquireLeaseRequest;
import xyz.block.ftl.v1.AcquireLeaseResponse;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;
import xyz.block.ftl.v1.ModuleContextRequest;
import xyz.block.ftl.v1.ModuleContextResponse;
import xyz.block.ftl.v1.PublishEventRequest;
import xyz.block.ftl.v1.PublishEventResponse;
import xyz.block.ftl.v1.VerbServiceGrpc;
import xyz.block.ftl.v1.schema.Ref;

@Singleton
@Startup
public class FTLController implements LeaseClient {
    private static final Logger log = Logger.getLogger(FTLController.class);
    final String moduleName;
    private StreamObserver<AcquireLeaseRequest> leaseClient;
    private final Deque<CompletableFuture<?>> leaseWaiters = new LinkedBlockingDeque<>();

    private Throwable currentError;
    private volatile ModuleContextResponse moduleContextResponse;
    private boolean waiters = false;
    private volatile boolean closed = false;

    final VerbServiceGrpc.VerbServiceStub verbService;
    final StreamObserver<ModuleContextResponse> moduleObserver = new StreamObserver<>() {
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
            if (!closed) {
                verbService.getModuleContext(ModuleContextRequest.newBuilder().setModule(moduleName).build(), moduleObserver);
            }
        }

        @Override
        public void onCompleted() {
            onError(new RuntimeException("connection closed"));
        }
    };

    @PreDestroy
    void shutdown() {

    }

    public FTLController(@ConfigProperty(name = "ftl.endpoint", defaultValue = "http://localhost:8892") URI uri,
            @ConfigProperty(name = "ftl.module.name") String moduleName) {
        this.moduleName = moduleName;
        var channelBuilder = ManagedChannelBuilder.forAddress(uri.getHost(), uri.getPort());
        if (uri.getScheme().equals("http")) {
            channelBuilder.usePlaintext();
        }
        var channel = channelBuilder.build();
        verbService = VerbServiceGrpc.newStub(channel);
        verbService.getModuleContext(ModuleContextRequest.newBuilder().setModule(moduleName).build(), moduleObserver);
        synchronized (this) {
            this.leaseClient = verbService.acquireLease(new StreamObserver<AcquireLeaseResponse>() {
                @Override
                public void onNext(AcquireLeaseResponse value) {
                    leaseWaiters.pop().complete(null);
                }

                @Override
                public void onError(Throwable t) {
                    synchronized (FTLController.this) {
                        while (!leaseWaiters.isEmpty()) {
                            leaseWaiters.pop().completeExceptionally(t);
                        }
                        if (!closed) {
                            leaseClient = verbService.acquireLease(this);
                        }
                    }
                }

                @Override
                public void onCompleted() {
                    //if we have any waiters error them out
                    //if we have not shut down we can try and connect again
                    onError(new RuntimeException("stream closed"));
                }
            });
        }
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
        verbService.publishEvent(PublishEventRequest.newBuilder()
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

    public void acquireLease(Duration duration, String... keys) throws LeaseFailedException {
        CompletableFuture<?> cf = new CompletableFuture<>();
        synchronized (this) {
            leaseWaiters.push(cf);
            leaseClient.onNext(AcquireLeaseRequest.newBuilder().setModule(moduleName)
                    .addAllKey(Arrays.asList(keys))
                    .setTtl(com.google.protobuf.Duration.newBuilder()
                            .setSeconds(duration.toSeconds()))
                    .build());
        }
        try {
            cf.get();
        } catch (Exception e) {
            throw new LeaseFailedException(e);
        }
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

}
