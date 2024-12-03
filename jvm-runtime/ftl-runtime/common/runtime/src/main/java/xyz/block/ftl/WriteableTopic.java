package xyz.block.ftl;

/**
 * A concrete definition of a topic. Extend this interface and annotate with {@code @TopicDefinition} to define a topic,
 * then inject this into verb methods to publish to the topic.
 *
 * @param <T> The type of the event to be published
 * @param <M> The type of the partition mapper
 */
public interface WriteableTopic<T, M extends TopicPartitionMapper<? super T>> extends ConsumableTopic {

    void publish(T object);
}
