package xyz.block.ftl.java.runtime.it;

import java.util.function.Function;

import jakarta.inject.Inject;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import ftl.echo.EchoClient;
import ftl.echo.EchoRequest;
import ftl.echo.EchoResponse;
import io.quarkus.test.common.QuarkusTestResource;
import io.quarkus.test.junit.QuarkusTest;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.VerbClientDefinition;
import xyz.block.ftl.VerbClientSink;
import xyz.block.ftl.java.test.FTLManaged;
import xyz.block.ftl.java.test.internal.FTLTestResource;
import xyz.block.ftl.java.test.internal.TestVerbServer;

@QuarkusTest
@QuarkusTestResource(FTLTestResource.class)
public class FtlJavaRuntimeResourceTest {

    @FTLManaged
    @Inject
    PublishVerbClient myVerbClient;

    @FTLManaged
    @Inject
    HelloClient helloClient;

    @Test
    public void testHelloEndpoint() {
        TestVerbServer.registerFakeVerb("echo", "echo", new Function<EchoRequest, EchoResponse>() {
            @Override
            public EchoResponse apply(EchoRequest s) {
                return new EchoResponse(s.getName());
            }
        });
        EchoClient echoClient = Mockito.mock(EchoClient.class);
        Mockito.when(echoClient.call(Mockito.any())).thenReturn(new EchoResponse().setMessage("Stuart"));
        Assertions.assertEquals("Hello Stuart", helloClient.call("Stuart"));
    }

    @Test
    @Disabled
    public void testTopic() {
        myVerbClient.call(new Person("Stuart", "Douglas"));
    }

    @VerbClientDefinition(name = "publish")
    interface PublishVerbClient extends VerbClientSink<Person> {
    }

    @VerbClientDefinition(name = "hello")
    interface HelloClient extends VerbClient<String, String> {
    }
}
