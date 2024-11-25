package xyz.block.ftl.java.test.http;

public class PostResponse {
    private boolean success;

    public boolean isSuccess() {
        return success;
    }

    public PostResponse setSuccess(boolean success) {
        this.success = success;
        return this;
    }
}