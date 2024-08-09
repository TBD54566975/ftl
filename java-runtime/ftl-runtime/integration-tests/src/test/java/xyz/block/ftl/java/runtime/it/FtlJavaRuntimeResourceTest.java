package xyz.block.ftl.java.runtime.it;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.is;

import ftl.echo.EchoClient;
import ftl.echo.EchoRequest;
import ftl.echo.EchoResponse;
import io.quarkus.test.InjectMock;
import jakarta.inject.Inject;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

import io.quarkus.test.junit.QuarkusTest;
import org.mockito.Mockito;

@QuarkusTest
public class FtlJavaRuntimeResourceTest {

    @Inject
    FtlJavaRuntimeResource resource;


    @Test
    public void testHelloEndpoint() {
        EchoResponse response = new EchoResponse();
        response.message = "Stuart";
        EchoClient echoClient = Mockito.mock(EchoClient.class);
        Mockito.when(echoClient.call(Mockito.any())).thenReturn(response);
        Assertions.assertEquals("Hello Stuart", resource.hello("Stuart", echoClient));
    }
    @Test
    public void testTopic() {
        MyTopic topic = Mockito.mock(MyTopic.class);
        resource.publish(new Person("Stuart", "Douglas"), topic);
        Mockito.verify(topic).publish(new Person("Stuart", "Douglas"));
    }
}
