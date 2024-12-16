package xyz.block.ftl.java.test.subscriber

import ftl.builtin.CatchRequest
import ftl.publisher.PubSubEvent
import ftl.publisher.TestTopicTopic
import ftl.publisher.Topic2Topic
import io.quarkus.logging.Log
import xyz.block.ftl.*
import java.util.concurrent.atomic.AtomicInteger

class Subscriber {
    @Subscription(topic = TestTopicTopic::class, from = FromOffset.BEGINNING)
    @Throws(
        Exception::class
    )
    fun consume(event: PubSubEvent) {
        Log.infof("Subscriber is consuming %s", event.time)
    }

    @Subscription(topic = TestTopicTopic::class, from = FromOffset.LATEST)
    fun consumeFromLatest(event: PubSubEvent) {
        Log.infof("Subscriber is consuming %s", event.time)
    }

    @Subscription(topic = Topic2Topic::class, from = FromOffset.BEGINNING)
    @Retry(count = 2, minBackoff = "1s", maxBackoff = "1s", catchVerb = "catch")
    fun consumeButFailAndRetry(event: PubSubEvent) {
        throw RuntimeException("always error: event " + event.time)
    }

    @Subscription(topic = Topic2Topic::class, from = FromOffset.BEGINNING)
    @Retry(count = 1, minBackoff = "1s", maxBackoff = "1s", catchVerb = "catchAny")
    fun consumeButFailAndCatchAny(event: PubSubEvent) {
        throw RuntimeException("always error: event " + event.time)
    }

    @Verb
    @VerbName("catch")
    fun catchVerb(req: CatchRequest<PubSubEvent?>) {
        if (!req.error.contains("always error: event")) {
            throw RuntimeException("unexpected error: " + req.error)
        }
        if (catchCount.incrementAndGet() == 1) {
            throw RuntimeException("catching error")
        }
    }

    @Verb
    fun catchAny(req: CatchRequest<Any>) {
        require("subscriber" == req.verb.module) { String.format("unexpected verb module: %s", req.verb.module) }
        require("consumeButFailAndCatchAny" == req.verb.name) {
            String.format(
                "unexpected verb name: %s",
                req.verb.name
            )
        }
        require("publisher.PubSubEvent" == req.requestType) {
            String.format(
                "unexpected request type: %s",
                req.requestType
            )
        }
        require(req.request is Map<*, *>) {
            String.format(
                "expected request to be a Map: %s",
                req.request.javaClass.name
            )
        }
        val request = req.request as Map<*, *>
        val time = request["time"]
        requireNotNull(time) { "expected request to have a time key" }
        require(time is String) { "expected request to have a time value of type string" }
    }

    companion object {
        private val catchCount = AtomicInteger()
    }
}
