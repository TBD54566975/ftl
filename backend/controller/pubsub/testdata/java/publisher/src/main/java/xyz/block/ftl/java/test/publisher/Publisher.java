package xyz.block.ftl.java.test.publisher;

import io.quarkus.logging.Log;
import xyz.block.ftl.Export;
import xyz.block.ftl.Topic;
import xyz.block.ftl.TopicDefinition;
import xyz.block.ftl.Verb;

public class Publisher {

    @Export
    @TopicDefinition("testTopic")
    interface TestTopic extends Topic<PubSubEvent> {

    }

    @Export
    @TopicDefinition("topic2")
    interface Topic2 extends Topic<PubSubEvent> {

    }

    @Verb
    void publishTen(TestTopic testTopic) throws Exception {
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
    void publishOneToTopic2(Topic2 topic2) throws Exception {
        var t = java.time.ZonedDateTime.now();
        Log.infof("Publishing %s", t);
        topic2.publish(new PubSubEvent().setTime(t));
    }
}
