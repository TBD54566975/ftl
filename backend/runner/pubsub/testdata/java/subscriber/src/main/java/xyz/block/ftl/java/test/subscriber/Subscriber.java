package xyz.block.ftl.java.test.subscriber;

import ftl.builtin.FailedEvent;
import ftl.publisher.PubSubEvent;
import ftl.publisher.TestTopicTopic;
import ftl.publisher.Topic2Topic;
import io.quarkus.logging.Log;
import xyz.block.ftl.FromOffset;
import xyz.block.ftl.Retry;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.Topic;
import xyz.block.ftl.TopicPartitionMapper;
import xyz.block.ftl.WriteableTopic;

class PartitionMapper implements TopicPartitionMapper<PubSubEvent> {
    public String getPartitionKey(PubSubEvent event) {
        return event.getTime().toString();
    }
}

public class Subscriber {

    @Subscription(topic = TestTopicTopic.class, from = FromOffset.BEGINNING)
    void consume(PubSubEvent event) throws Exception {
        Log.infof("consume: %s", event.getTime());
    }

    @Subscription(topic = TestTopicTopic.class, from = FromOffset.LATEST)
    void consumeFromLatest(PubSubEvent event) throws Exception {
        Log.infof("consumeFromLatest: %s", event.getTime());
    }

    // Java requires the topic to be explicitly defined as an interface for consuming to work
    @Topic("consumeButFailAndRetryFailed")
    interface ConsumeButFailAndRetryFailedTopic extends WriteableTopic<PubSubEvent, PartitionMapper> {

    }

    @Subscription(topic = Topic2Topic.class, from = FromOffset.BEGINNING, deadLetter = true)
    @Retry(count = 2, minBackoff = "1s", maxBackoff = "1s")
    public void consumeButFailAndRetry(PubSubEvent event) {
        throw new RuntimeException("always error: event " + event.getTime());
    }

    @Subscription(topic = ConsumeButFailAndRetryFailedTopic.class, from = FromOffset.BEGINNING)
    public void consumeFromDeadLetter(builtin.FailedEvent<PubSubEvent> event) {
        throw new RuntimeException("always error: event " + event.getEvent().getTime());
    }
}
