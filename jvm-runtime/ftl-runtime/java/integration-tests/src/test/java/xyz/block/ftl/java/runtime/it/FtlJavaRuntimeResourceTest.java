package xyz.block.ftl.java.runtime.it;

import java.nio.charset.StandardCharsets;
import java.util.function.Function;

import jakarta.inject.Inject;

import org.hamcrest.Matchers;
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
import io.restassured.RestAssured;
import io.restassured.http.ContentType;
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
        RestAssured.with().body(new Person("Stuart", "Douglas"))
                .contentType(ContentType.JSON)
                .post("/test/post")
                .then()
                .statusCode(200)
                .body(Matchers.equalTo("Hello Stuart Douglas"));
    }

    @Test
    public void testHttpBytes() {

        RestAssured.with().body("Stuart Douglas".getBytes(java.nio.charset.StandardCharsets.UTF_8))
                .contentType(ContentType.JSON)
                .post("/test/bytes")
                .then()
                .statusCode(200)
                .body(Matchers.equalTo("Hello Stuart Douglas"));
    }

    @VerbClientDefinition(name = "publish")
    interface PublishVerbClient extends VerbClientSink<Person> {
    }

    @VerbClientDefinition(name = "bytes")
    interface BytesClient extends VerbClient<byte[], byte[]> {
    }

    @VerbClientDefinition(name = "bytesHttp")
    interface BytesHTTPClient extends VerbClient<HttpRequest<byte[]>, HttpResponse<String, String>> {
    }
}
