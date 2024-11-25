package xyz.block.ftl.java.test.http;

import java.util.List;

import jakarta.ws.rs.DELETE;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.PUT;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;

import jakarta.ws.rs.core.MediaType;
import org.jboss.resteasy.reactive.ResponseHeader;
import org.jboss.resteasy.reactive.ResponseStatus;
import org.jboss.resteasy.reactive.RestPath;
import org.jboss.resteasy.reactive.RestQuery;

@Path("/")
public class TestHTTP {

    @GET
    @Path("/users/{userId}/posts/{postId}")
    @ResponseHeader(name = "Get", value = "Header from FTL")
    public GetResponse get(@RestPath int userId, @RestPath int postId) {
        return new GetResponse()
                .setMsg(String.format("UserID: %s, PostID: %s", userId, postId))
                .setNested(new Nested().setGoodStuff("This is good stuff"));
    }


    @GET
    @Path("/getquery")
    @ResponseHeader(name = "Get", value = "Header from FTL")
    public GetResponse getquery(@RestQuery int userId, @RestQuery int postId) {
        return new GetResponse()
                .setMsg(String.format("UserID: %s, PostID: %s", userId, postId))
                .setNested(new Nested().setGoodStuff("This is good stuff"));
    }

    @POST
    @Path("/users")
    @ResponseStatus(201)
    @ResponseHeader(name = "Post", value = "Header from FTL")
    public PostResponse post(PostRequest req) {
        return new PostResponse().setSuccess(true);
    }

    @PUT
    @Path("/users/{userId}")
    @ResponseHeader(name = "Put", value = "Header from FTL")
    public PutResponse put(PutRequest req) {
        return new PutResponse();
    }

    @DELETE
    @Path("/users/{userId}")
    @ResponseHeader(name = "Delete", value = "Header from FTL")
    @ResponseStatus(200)
    public DeleteResponse delete(@RestPath String userId) {
        System.out.println("delete");
        return new DeleteResponse();
    }

    @GET
    @Path("/queryparams")
    public String query(@RestQuery String foo) {
        return foo == null ? "No value" : foo;
    }

    @GET
    @Path("/html")
    @Produces("text/html; charset=utf-8")
    public String html() {
        return "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>";
    }

    @POST
    @Path("/bytes")
    public byte[] bytes(byte[] b) {
        return b;
    }

    @GET
    @Path("/empty")
    @ResponseStatus(200)
    public void empty() {
    }

    @POST
    @Path("/string")
    public String string(String val) {
        return val;
    }

    @POST
    @Path("/int")
    @Produces(MediaType.APPLICATION_JSON)
    public int intMethod(int val) {
        return val;
    }

    @POST
    @Path("/float")
    @Produces(MediaType.APPLICATION_JSON)
    public float floatVerb(float val) {
        return val;
    }

    @POST
    @Path("/bool")
    @Produces(MediaType.APPLICATION_JSON)
    public boolean bool(boolean val) {
        return val;
    }

    @GET
    @Path("/error")
    public String error() {
        throw new RuntimeException("Error from FTL");
    }

    @POST
    @Path("/array/string")
    public String[] arrayString(String[] array) {
        return array;
    }

    @POST
    @Path("/array/data")
    public List<ArrayType> arrayData(List<ArrayType> array) {
        return array;
    }

}
