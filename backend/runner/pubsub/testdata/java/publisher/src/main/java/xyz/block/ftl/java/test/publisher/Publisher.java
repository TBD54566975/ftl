package xyz.block.ftl.java.test.publisher;

import io.quarkus.logging.Log;
import xyz.block.ftl.Export;
import xyz.block.ftl.FromOffset;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.Topic;
import xyz.block.ftl.TopicPartitionMapper;
import xyz.block.ftl.Verb;
import xyz.block.ftl.WriteableTopic;

class PartitionMapper implements TopicPartitionMapper<PubSubEvent> {
    public String getPartitionKey(PubSubEvent event) {
        return event.getTime().toString();
    }
}

public class Publisher {

    @Export
    @Topic("testTopic")
    interface TestTopic extends WriteableTopic<PubSubEvent, PartitionMapper> {

    }

    @Topic("localTopic")
    interface LocalTopic extends WriteableTopic<PubSubEvent, PartitionMapper> {

    }

    @Export
    @Topic("topic2")
    interface Topic2 extends WriteableTopic<PubSubEvent, PartitionMapper> {

    }

    @Verb
    void publishTen(LocalTopic testTopic) throws Exception {
        for (var i = 0; i < 10; ++i) {
            var t = java.time.ZonedDateTime.now();
            Log.infof("Publishing %s", t);
            testTopic.publish(new PubSubEvent().setTime(t));
        }
    }

    @Verb
    void publishOne(TestTopic testTopic) throws Exception {
        var t = java.time.ZonedDateTime.now();
        Log.infof("Publishing %s", t);
        testTopic.publish(new PubSubEvent().setTime(t));
    }

    @Verb
    void publishOneToTopic2(HaystackRequest req, Topic2 topic2) throws Exception {
        var t = java.time.ZonedDateTime.now();
        Log.infof("Publishing %s", t);
        topic2.publish(new PubSubEvent().setTime(t).setHaystack(req.getHaystack()));
    }

    @Subscription(topic = LocalTopic.class, from = FromOffset.LATEST)
    public void local(TestTopic testTopic, PubSubEvent event) {
        testTopic.publish(event);
    }
}
