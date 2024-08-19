package xyz.block.ftl.java.test.http;

import com.fasterxml.jackson.annotation.JsonAlias;

public class PostRequest {
    @JsonAlias("user_id")
    private int userId;
    private int postId;

    public int getUserId() {
        return userId;
    }

    public void setUserId(int userId) {
        this.userId = userId;
    }

    public int getPostId() {
        return postId;
    }

    public void setPostId(int postId) {
        this.postId = postId;
    }
}