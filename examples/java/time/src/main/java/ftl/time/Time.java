package ftl.time;

import java.time.OffsetDateTime;

import io.opentelemetry.api.metrics.LongCounter;
import io.opentelemetry.api.metrics.Meter;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Time {
    final LongCounter counter;

    public Time(Meter meter) {
        counter = meter.counterBuilder("time.invocations")
                .setDescription("The number of time invocations")
                .setUnit("invocations")
                .build();
    }

    @Verb
    @Export
    public TimeResponse time() {
        counter.add(1);
        return new TimeResponse(OffsetDateTime.now());
    }
}
