package xyz.block.ftl.deployment;

import java.util.HashMap;
import java.util.Map;

import org.jboss.jandex.DotName;
import org.jboss.jandex.Type;

import io.quarkus.builder.item.SimpleBuildItem;

public final class TopicsBuildItem extends SimpleBuildItem {

    final Map<DotName, DiscoveredTopic> topics;

    public TopicsBuildItem(Map<DotName, DiscoveredTopic> topics) {
        this.topics = new HashMap<>(topics);
    }

    public Map<DotName, DiscoveredTopic> getTopics() {
        return topics;
    }

    public record DiscoveredTopic(String topicName, String generatedProducer, Type eventType, boolean exported,
            String interfaceName) {

    }
}
