package xyz.block.ftl.runtime;

import com.google.protobuf.ByteString;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import io.quarkus.logging.Log;
import io.quarkus.runtime.Startup;
import io.quarkus.runtime.annotations.ConfigItem;
import jakarta.inject.Singleton;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import xyz.block.ftl.v1.ModuleContextRequest;
import xyz.block.ftl.v1.ModuleContextResponse;
import xyz.block.ftl.v1.VerbServiceGrpc;

import java.net.URI;

@Singleton
@Startup
public class FTL {
    final String moduleName;

    private Throwable currentError;
    private volatile ModuleContextResponse moduleContextResponse;
    private boolean waiters = false;

    final VerbServiceGrpc.VerbServiceStub verbService;
    final StreamObserver<ModuleContextResponse> moduleObserver = new StreamObserver<>() {
        @Override
        public void onNext(ModuleContextResponse moduleContextResponse) {
            synchronized (this) {
                currentError = null;
                FTL.this.moduleContextResponse = moduleContextResponse;
                if (waiters) {
                    this.notifyAll();
                    waiters = false;
                }
            }

        }

        @Override
        public void onError(Throwable throwable) {
            Log.error("GRPC connection error", throwable);
            synchronized (this) {
                currentError = throwable;
                if (waiters) {
                    this.notifyAll();
                    waiters = false;
                }
            }
        }

        @Override
        public void onCompleted() {
            verbService.getModuleContext(ModuleContextRequest.newBuilder().setModule(moduleName).build(), moduleObserver);
        }
    };

    public FTL(@ConfigProperty(name = "ftl.endpoint", defaultValue = "http://localhost:8892") URI uri, @ConfigItem(name = "ftl.module.name") String moduleName) {
        this.moduleName = moduleName;
        var channelBuilder = ManagedChannelBuilder.forAddress(uri.getHost(), uri.getPort());
        if (uri.getScheme().equals("http")) {
            channelBuilder.usePlaintext();
        }
        var channel = channelBuilder.build();
        verbService = VerbServiceGrpc.newStub(channel);
        verbService.getModuleContext(ModuleContextRequest.newBuilder().setModule(moduleName).build(), moduleObserver);
    }

    public byte[] getSecret(String secretName) {
        var context = getModuleContext();
        if (context.containsSecrets(secretName)) {
            return context.getSecretsMap().get(secretName).toByteArray();
        }
        throw new RuntimeException("Secret not found: " + secretName);
    }

    private ModuleContextResponse getModuleContext() {
        var moduleContext = moduleContextResponse;
        if (moduleContext != null) {
            return moduleContext;
        }
        synchronized (moduleObserver) {
            for (; ; ) {
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
