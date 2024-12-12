package xyz.block.ftl.java.test.subscriber;

import java.util.concurrent.atomic.AtomicInteger;

import ftl.publisher.PubSubEvent;
import ftl.publisher.TestTopicTopic;
import ftl.publisher.Topic2Topic;
import io.quarkus.logging.Log;
import xyz.block.ftl.FromOffset;
import xyz.block.ftl.Retry;
import xyz.block.ftl.Subscription;

public class Subscriber {

    private static final AtomicInteger catchCount = new AtomicInteger();

    @Subscription(topic = TestTopicTopic.class, from = FromOffset.BEGINNING)
    void consume(PubSubEvent event) throws Exception {
        Log.infof("Subscriber is consuming %s", event.getTime());
    }

    @Subscription(topic = Topic2Topic.class, from = FromOffset.BEGINNING)
    @Retry(count = 2, minBackoff = "1s", maxBackoff = "1s")
    public void consumeButFailAndRetry(PubSubEvent event) {
        throw new RuntimeException("always error: event " + event.getTime());
    }
}
