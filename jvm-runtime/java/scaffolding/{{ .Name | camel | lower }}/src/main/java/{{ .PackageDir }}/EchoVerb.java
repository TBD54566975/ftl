package {{ .Group }};

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class EchoVerb {

    @Export
    @Verb
    public String echo(String request) {
        return "Hello, " + request + "!";
    }
}
