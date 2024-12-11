package xyz.block.ftl.runtime;

import java.io.Closeable;
import java.net.URI;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicInteger;

import org.jboss.logging.Logger;

import com.google.protobuf.ByteString;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.deployment.v1.DeploymentServiceGrpc;
import xyz.block.ftl.deployment.v1.GetDeploymentContextRequest;
import xyz.block.ftl.deployment.v1.GetDeploymentContextResponse;
import xyz.block.ftl.lease.v1.AcquireLeaseRequest;
import xyz.block.ftl.lease.v1.AcquireLeaseResponse;
import xyz.block.ftl.lease.v1.LeaseServiceGrpc;
import xyz.block.ftl.publish.v1.PublishEventRequest;
import xyz.block.ftl.publish.v1.PublishEventResponse;
import xyz.block.ftl.publish.v1.PublishServiceGrpc;
import xyz.block.ftl.schema.v1.Ref;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;
import xyz.block.ftl.v1.VerbServiceGrpc;

class FTLRunnerConnection implements Closeable {
    private static final Logger log = Logger.getLogger(FTLRunnerConnection.class);
    final String moduleName;
    final String deploymentName;
    private final ManagedChannel channel;
    private final String endpoint;

    private Throwable currentError;
    private volatile GetDeploymentContextResponse moduleContextResponse;
    private boolean waiters = false;

    final VerbServiceGrpc.VerbServiceStub verbService;
    final DeploymentServiceGrpc.DeploymentServiceStub deploymentService;
    final LeaseServiceGrpc.LeaseServiceStub leaseService;
    final PublishServiceGrpc.PublishServiceStub publishService;
    final StreamObserver<GetDeploymentContextResponse> moduleObserver = new ModuleObserver();

    FTLRunnerConnection(final String endpoint, final String deploymentName, final String moduleName) {
        var uri = URI.create(endpoint);
        this.moduleName = moduleName;
        var channelBuilder = ManagedChannelBuilder.forAddress(uri.getHost(), uri.getPort());
        if (uri.getScheme().equals("http")) {
            channelBuilder.usePlaintext();
        }
        this.channel = channelBuilder.build();
        this.deploymentName = deploymentName;
        deploymentService = DeploymentServiceGrpc.newStub(channel);
        deploymentService.getDeploymentContext(GetDeploymentContextRequest.newBuilder().setDeployment(deploymentName).build(),
                moduleObserver);
        verbService = VerbServiceGrpc.newStub(channel);
        publishService = PublishServiceGrpc.newStub(channel);
        leaseService = LeaseServiceGrpc.newStub(channel);
        this.endpoint = endpoint;
    }

    public String getEndpoint() {
        return endpoint;
    }

    byte[] getSecret(String secretName) {
        var context = getDeploymentContext();
        if (context.containsSecrets(secretName)) {
            return context.getSecretsMap().get(secretName).toByteArray();
        }
        throw new RuntimeException("Secret not found: " + secretName);
    }

    byte[] getConfig(String secretName) {
        var context = getDeploymentContext();
        if (context.containsConfigs(secretName)) {
            return context.getConfigsMap().get(secretName).toByteArray();
        }
        throw new RuntimeException("Config not found: " + secretName);
    }

    byte[] callVerb(String name, String module, byte[] payload) {
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

    void publishEvent(String topic, String callingVerbName, byte[] event, String key) {
        CompletableFuture<?> cf = new CompletableFuture<>();
        publishService.publishEvent(PublishEventRequest.newBuilder()
                .setCaller(callingVerbName).setBody(ByteString.copyFrom(event))
                .setTopic(Ref.newBuilder().setModule(moduleName).setName(topic).build())
                .setKey(key).build(),
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
        var client = leaseService.acquireLease(new StreamObserver<AcquireLeaseResponse>() {
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
        List<String> realKeys = new ArrayList<>();
        realKeys.add("module");
        realKeys.add(moduleName);
        realKeys.addAll(Arrays.asList(keys));
        client.onNext(AcquireLeaseRequest.newBuilder()
                .addAllKey(realKeys)
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

    GetDeploymentContextResponse getDeploymentContext() {
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

    @Override
    public void close() {
        channel.shutdown();
    }

    private class ModuleObserver implements StreamObserver<GetDeploymentContextResponse> {

        final AtomicInteger failCount = new AtomicInteger();

        @Override
        public void onNext(GetDeploymentContextResponse moduleContextResponse) {
            synchronized (this) {
                currentError = null;
                FTLRunnerConnection.this.moduleContextResponse = moduleContextResponse;
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
                deploymentService.getDeploymentContext(
                        GetDeploymentContextRequest.newBuilder().setDeployment(deploymentName).build(),
                        moduleObserver);
            }
        }

        @Override
        public void onCompleted() {
            onError(new RuntimeException("connection closed"));
        }
    }

}
