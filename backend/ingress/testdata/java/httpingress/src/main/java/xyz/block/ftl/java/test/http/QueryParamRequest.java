package xyz.block.ftl.java.test.http;

import org.jetbrains.annotations.Nullable;

public class QueryParamRequest {

    @Nullable
    String foo;

    public @Nullable String getFoo() {
        return foo;
    }

    public QueryParamRequest setFoo(@Nullable String foo) {
        this.foo = foo;
        return this;
    }
}
