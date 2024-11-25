package xyz.block.ftl.java.test.http;

import com.fasterxml.jackson.annotation.JsonAlias;

public class Nested {

    @JsonAlias("good_stuff")
    private String goodStuff;

    public String getGoodStuff() {
        return goodStuff;
    }

    public Nested setGoodStuff(String goodStuff) {
        this.goodStuff = goodStuff;
        return this;
    }
}
