package xyz.block.ftl.java.test.publisher;

import java.time.ZonedDateTime;

public class PubSubEvent {

    private ZonedDateTime time;

    public ZonedDateTime getTime() {
        return time;
    }

    public PubSubEvent setTime(ZonedDateTime time) {
        this.time = time;
        return this;
    }
}
