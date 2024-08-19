package xyz.block.ftl.java.test.subscriber;

import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import ftl.builtin.CatchRequest;
import ftl.publisher.PubSubEvent;
import ftl.publisher.TestTopicSubscription;
import io.quarkus.logging.Log;
import xyz.block.ftl.Retry;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.Verb;
import xyz.block.ftl.VerbName;

public class Subscriber {

    private static final AtomicInteger catchCount = new AtomicInteger();

    @TestTopicSubscription
    void consume(PubSubEvent event) throws Exception {
        Log.infof("Subscriber is consuming %s", event.getTime());
    }

    @Subscription(topic = "topic2", module = "publisher", name = "doomedSubscription")
    @Retry(count = 2, minBackoff = "1s", maxBackoff = "1s", catchVerb = "catch")
    public void consumeButFailAndRetry(PubSubEvent event) {
        throw new RuntimeException("always error: event " + event.getTime());
    }

    @Subscription(topic = "topic2", module = "publisher", name = "doomedSubscription2")
    @Retry(count = 1, minBackoff = "1s", maxBackoff = "1s", catchVerb = "catchAny")
    public void consumeButFailAndCatchAny(PubSubEvent event) {
        throw new RuntimeException("always error: event " + event.getTime());
    }

    @Verb
    @VerbName("catch")
    public void catchVerb(CatchRequest<PubSubEvent> req) {
        if (!req.getError().contains("always error: event")) {
            throw new RuntimeException("unexpected error: " + req.getError());
        }
        if (catchCount.incrementAndGet() == 1) {
            throw new RuntimeException("catching error");
        }
    }

    @Verb
    public void catchAny(CatchRequest<Object> req) {
        if (!"subscriber".equals(req.getVerb().getModule())) {
            throw new IllegalArgumentException(String.format("unexpected verb module: %s", req.getVerb().getModule()));
        }
        if (!"consumeButFailAndCatchAny".equals(req.getVerb().getName())) {
            throw new IllegalArgumentException(String.format("unexpected verb name: %s", req.getVerb().getName()));
        }
        if (!"publisher.PubSubEvent".equals(req.getRequestType())) {
            throw new IllegalArgumentException(String.format("unexpected request type: %s", req.getRequestType()));
        }
        if (!(req.getRequest()instanceof Map<?,?>)) {
            throw new IllegalArgumentException(
                    String.format("expected request to be a Map: %s", req.getRequest().getClass().getName()));
        }
        var request = (Map<?, ?>) req.getRequest();
        var time = request.get("time");
        if (time == null) {
            throw new IllegalArgumentException("expected request to have a time key");
        }
        if (!(time instanceof String)) {
            throw new IllegalArgumentException("expected request to have a time value of type string");
        }
    }

}
