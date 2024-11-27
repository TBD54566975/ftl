package xyz.block.ftl.java.test.publisher

import io.quarkus.logging.Log
import xyz.block.ftl.*
import java.time.ZonedDateTime

class Publisher {
    @Export
    @Topic("testTopic")
    interface TestTopic : WriteableTopic<PubSubEvent?>

    @Topic("localTopic")
    interface LocalTopic : WriteableTopic<PubSubEvent?>

    @Export
    @Topic("topic2")
    interface Topic2 : WriteableTopic<PubSubEvent?>

    @Verb
    @Throws(Exception::class)
    fun publishTen(testTopic: LocalTopic) {
        for (i in 0..9) {
            val t = ZonedDateTime.now()
            Log.infof("Publishing %s", t)
            testTopic.publish(PubSubEvent(t))
        }
    }

    @Verb
    @Throws(Exception::class)
    fun publishOne(testTopic: TestTopic) {
        val t = ZonedDateTime.now()
        Log.infof("Publishing %s", t)
        testTopic.publish(PubSubEvent(t))
    }

    @Verb
    @Throws(Exception::class)
    fun publishOneToTopic2(topic2: Topic2) {
        val t = ZonedDateTime.now()
        Log.infof("Publishing %s", t)
        topic2.publish(PubSubEvent(t))
    }

    @Subscription(topic = LocalTopic::class, from = FromOffset.LATEST)
    fun local(testTopic: TestTopic, event: PubSubEvent) {
        testTopic.publish(event)
    }
}
