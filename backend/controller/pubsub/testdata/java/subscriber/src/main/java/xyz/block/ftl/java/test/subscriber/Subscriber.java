package xyz.block.ftl.java.test.subscriber;

import ftl.builtin.CatchRequest;
import ftl.publisher.PubSubEvent;
import ftl.publisher.TestTopicSubscription;
import ftl.publisher.Topic2Subscription;
import io.quarkus.logging.Log;
import xyz.block.ftl.Retry;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.TopicDefinition;
import xyz.block.ftl.Verb;
import xyz.block.ftl.VerbName;

import java.util.concurrent.atomic.AtomicInteger;

public class Subscriber {

    private static final AtomicInteger catchCount = new AtomicInteger();
    @TestTopicSubscription
    void consume(PubSubEvent event) throws Exception {
        Log.infof("Subscriber is consuming %s", event.getTime());
    }

    @Subscription(
            topic = "topic2",
            module = "publisher",
            name = "doomedSubscription"
    )
    @Retry(count = 2, minBackoff = "1s", maxBackoff = "1s", catchVerb = "catch")
    public void consumeButFailAndRetry(PubSubEvent event) {
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
}

