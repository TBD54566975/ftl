package xyz.block.ftl.java.test.publisher

import io.quarkus.logging.Log
import xyz.block.ftl.*
import java.time.ZonedDateTime

class PartitionMapper : TopicPartitionMapper<PubSubEvent> {
    override fun getPartitionKey(event: PubSubEvent): String {
        return event.time.toString()
    }
}

class Publisher {
    @Export
    @Topic("testTopic")
    interface TestTopic : WriteableTopic<PubSubEvent, PartitionMapper>

    @Topic("localTopic")
    interface LocalTopic : WriteableTopic<PubSubEvent, PartitionMapper>

    @Export
    @Topic("topic2")
    interface Topic2 : WriteableTopic<PubSubEvent, PartitionMapper>

    @Verb
    @Throws(Exception::class)
    fun publishTen(testTopic: TestTopic) {
        for (i in 0..9) {
            val t = ZonedDateTime.now()
            Log.infof("Publishing to testTopic: %s", t)
            testTopic.publish(PubSubEvent(t))
        }
    }

    @Verb
    @Throws(Exception::class)
    fun publishTenLocal(localTopic: LocalTopic) {
        for (i in 0..9) {
            val t = ZonedDateTime.now()
            Log.infof("Publishing to localTopic: %s", t)
            localTopic.publish(PubSubEvent(t))
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
    fun local(event: PubSubEvent) {
        Log.infof("Consuing from local %s", event.time)
    }
}
