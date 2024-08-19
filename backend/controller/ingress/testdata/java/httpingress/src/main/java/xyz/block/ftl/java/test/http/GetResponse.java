package xyz.block.ftl.java.test.http;

public class GetResponse {

    private Nested nested;

    private String msg;

    public Nested getNested() {
        return nested;
    }

    public GetResponse setNested(Nested nested) {
        this.nested = nested;
        return this;

    }

    public String getMsg() {
        return msg;
    }

    public GetResponse setMsg(String msg) {
        this.msg = msg;
        return this;
    }
}
