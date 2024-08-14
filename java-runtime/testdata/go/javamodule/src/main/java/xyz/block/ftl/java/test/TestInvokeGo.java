package xyz.block.ftl.java.test;

import ftl.gomodule.EmptyVerbClient;
import ftl.gomodule.ErrorEmptyVerbClient;
import ftl.gomodule.SinkVerbClient;
import ftl.gomodule.SourceVerbClient;
import ftl.gomodule.TimeClient;
import ftl.gomodule.TimeRequest;
import ftl.gomodule.TimeResponse;
import org.jetbrains.annotations.NotNull;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class TestInvokeGo {

    @Export
    @Verb
    public void emptyVerb(EmptyVerbClient emptyVerbClient) {
        emptyVerbClient.call();
    }

    @Export
    @Verb
    public void sinkVerb(String input, SinkVerbClient sinkVerbClient) {
        sinkVerbClient.call(input);
    }

    @Export
    @Verb
    public String sourceVerb(SourceVerbClient sourceVerbClient) {
        return sourceVerbClient.call();
    }
    @Export
    @Verb
    public void errorEmptyVerb(ErrorEmptyVerbClient client) {
         client.call();
    }

    @Export
    @Verb
    public @NotNull TimeResponse timeVerb(TimeClient client) {
        return client.call(new TimeRequest());
    }

}
