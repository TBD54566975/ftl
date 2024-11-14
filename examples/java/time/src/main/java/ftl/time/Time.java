package ftl.time;

import java.time.OffsetDateTime;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Time {

    @Verb
    @Export
    public TimeResponse time() {
        return new TimeResponse(OffsetDateTime.now(), "foo");
    }
}
