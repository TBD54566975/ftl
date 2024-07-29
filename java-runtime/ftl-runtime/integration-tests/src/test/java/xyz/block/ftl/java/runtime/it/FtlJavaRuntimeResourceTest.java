package xyz.block.ftl.java.runtime.it;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.is;

import org.junit.jupiter.api.Test;

import io.quarkus.test.junit.QuarkusTest;

@QuarkusTest
public class FtlJavaRuntimeResourceTest {

    @Test
    public void testHelloEndpoint() {
        given()
                .when().get("/ftl-java-runtime")
                .then()
                .statusCode(200)
                .body(is("Hello ftl-java-runtime"));
    }
}
