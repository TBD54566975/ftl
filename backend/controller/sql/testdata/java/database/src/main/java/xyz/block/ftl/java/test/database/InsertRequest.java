package xyz.block.ftl.java.test.database;

public class InsertRequest {
    private String data;
    private int id;

    public int getId() {
        return id;
    }

    public InsertRequest setId(int id) {
        this.id = id;
        return this;
    }

    public String getData() {
        return data;
    }

    public InsertRequest setData(String data) {
        this.data = data;
        return this;
    }
}
