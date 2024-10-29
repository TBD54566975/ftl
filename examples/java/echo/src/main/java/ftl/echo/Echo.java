package ftl.echo;

import ftl.time.TimeClient;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Echo {

    @Export
    @Verb
    public EchoResponse echo(EchoRequest req, TimeClient time) {
        var response = time.time();
        return new EchoResponse("Hello, " + req.name().orElse("anonymous") + "! The time is " + response.toString() + ".");
    }
}
