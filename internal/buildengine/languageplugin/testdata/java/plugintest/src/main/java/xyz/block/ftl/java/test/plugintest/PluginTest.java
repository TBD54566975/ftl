package xyz.block.ftl.java.test.plugintest;

import xyz.block.ftl.Verb;

// uncommentForDependency:import ftl.dependable.Data;

public class PluginTest {
    // uncommentForDependency: Data data;
    @Verb
    public EchoResponse verbaabbcc(EchoRequest req) {

        EchoResponse rr = new EchoResponse();
        rr.message = String.format("Hello, %s!", req.name.orElse(("anonymous")));
        return rr;
    }

}
