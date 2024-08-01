package xyz.block.ftl.deployment;

import com.squareup.javapoet.JavaFile;
import com.squareup.javapoet.MethodSpec;
import com.squareup.javapoet.TypeSpec;
import io.grpc.ManagedChannelBuilder;
import io.quarkus.bootstrap.model.ApplicationModel;
import io.quarkus.bootstrap.prebuild.CodeGenException;
import io.quarkus.deployment.CodeGenContext;
import io.quarkus.deployment.CodeGenProvider;
import org.eclipse.microprofile.config.Config;
import xyz.block.ftl.v1.ControllerServiceGrpc;
import xyz.block.ftl.v1.GetSchemaRequest;

import javax.lang.model.element.Modifier;
import java.io.IOException;
import java.net.URI;
import java.nio.file.Path;
import java.util.Map;

public class FTLCodeGenerator implements CodeGenProvider {

    public static final String CLIENT = "Client";
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
        return "ftl";
    }

    @Override
    public boolean trigger(CodeGenContext context) throws CodeGenException {

        String endpoint = System.getenv("FTL_ENDPOINT");
        if (endpoint == null) {
            endpoint = "http://localhost:8892";
        }
        URI uri = URI.create(endpoint);

        var channelBuilder = ManagedChannelBuilder.forAddress(uri.getHost(), uri.getPort());
        if (uri.getScheme().equals("http")) {
            channelBuilder.usePlaintext();
        }

        var channel = channelBuilder.build();
        var blockingStub = ControllerServiceGrpc.newBlockingStub(channel);
        var schemaResponse = blockingStub.getSchema(GetSchemaRequest.newBuilder().build());
        for (var module : schemaResponse.getSchema().getModulesList()) {
            for (var decl : module.getDeclsList()) {
                var verb = decl.getVerb();
                if (verb != null && !verb.getName().isEmpty()) {
                    try {

                        MethodSpec call = MethodSpec.methodBuilder("call")
                                .addModifiers(Modifier.PUBLIC, Modifier.ABSTRACT)
                                .returns(void.class)
                                .addParameter(String[].class, "args")
                                .build();

                        TypeSpec helloWorld = TypeSpec.interfaceBuilder(className(verb.getName()) + CLIENT)
                                .addModifiers(Modifier.PUBLIC)
                                .addMethod(call)
                                .build();

                        JavaFile javaFile = JavaFile.builder(module.getName(), helloWorld)
                                .build();

                        javaFile.writeTo(context.outDir());

                    } catch (IOException e) {
                        throw new RuntimeException(e);
                    }
                }
            }
        }

        return true;
    }

    @Override
    public boolean shouldRun(Path sourceDir, Config config) {
        return true;
    }


    String className(String in ) {
        return Character.toUpperCase(in.charAt(0)) + in.substring(1);
    }
}
