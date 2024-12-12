package xyz.block.ftl.java.test.publisher;

import java.time.ZonedDateTime;

public class PubSubEvent {

    private ZonedDateTime time;
    private String haystack;

    public ZonedDateTime getTime() {
        return time;
    }

    public PubSubEvent setTime(ZonedDateTime time) {
        this.time = time;
        return this;
    }

    public String getHaystack() {
        return haystack;
    }

    public PubSubEvent setHaystack(String haystack) {
        this.haystack = haystack;
        return this;
    }
}
