package xyz.block.ftl.java.runtime.it;

import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.HashMap;
import java.util.function.Function;

import jakarta.inject.Inject;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import ftl.builtin.HttpRequest;
import ftl.builtin.HttpResponse;
import ftl.echo.EchoClient;
import ftl.echo.EchoRequest;
import ftl.echo.EchoResponse;
import io.quarkus.test.common.WithTestResource;
import io.quarkus.test.junit.QuarkusTest;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.VerbClientDefinition;
import xyz.block.ftl.VerbClientSink;
import xyz.block.ftl.java.test.FTLManaged;
import xyz.block.ftl.java.test.internal.FTLTestResource;
import xyz.block.ftl.java.test.internal.TestVerbServer;

@QuarkusTest
@WithTestResource(FTLTestResource.class)
public class FtlJavaRuntimeResourceTest {

    @FTLManaged
    @Inject
    PublishVerbClient myVerbClient;

    @FTLManaged
    @Inject
    HelloClient helloClient;

    @FTLManaged
    @Inject
    BytesClient bytesClient;

    @FTLManaged
    @Inject
    PostClient postClient;

    @FTLManaged
    @Inject
    BytesHTTPClient bytesHttpClient;

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

    @Test
    public void testBytesSerialization() {
        Assertions.assertArrayEquals(new byte[] { 1, 2 }, bytesClient.call(new byte[] { 1, 2 }));
    }

    @Test
    public void testHttpPost() {
        HttpRequest<Person> request = new HttpRequest<Person>()
                .setMethod("POST")
                .setPath("/test/post")
                .setQuery(new HashMap<>())
                .setPathParameters(new HashMap<>())
                .setHeaders(new HashMap<>())
                .setBody(new Person("Stuart", "Douglas"));
        HttpResponse<String, String> response = postClient.call(request);
        Assertions.assertEquals("Hello Stuart Douglas", response.getBody());
    }

    @Test
    public void testHttpBytes() {
        HttpRequest<byte[]> request = new HttpRequest<byte[]>()
                .setMethod("POST")
                .setPath("/test/bytes")
                .setQuery(new HashMap<>())
                .setPathParameters(new HashMap<>())
                .setHeaders(new HashMap<>())
                .setBody("Stuart Douglas".getBytes(java.nio.charset.StandardCharsets.UTF_8));
        HttpResponse<String, String> response = bytesHttpClient.call(request);
        Assertions.assertArrayEquals("Hello Stuart Douglas".getBytes(StandardCharsets.UTF_8),
                Base64.getDecoder().decode(response.getBody()));
    }

    @VerbClientDefinition(name = "publish")
    interface PublishVerbClient extends VerbClientSink<Person> {
    }

    @VerbClientDefinition(name = "hello")
    interface HelloClient extends VerbClient<String, String> {
    }

    @VerbClientDefinition(name = "bytes")
    interface BytesClient extends VerbClient<byte[], byte[]> {
    }

    @VerbClientDefinition(name = "post")
    interface PostClient extends VerbClient<HttpRequest<Person>, HttpResponse<String, String>> {
    }

    @VerbClientDefinition(name = "bytesHttp")
    interface BytesHTTPClient extends VerbClient<HttpRequest<byte[]>, HttpResponse<String, String>> {
    }
}
