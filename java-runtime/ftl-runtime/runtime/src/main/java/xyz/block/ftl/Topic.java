package xyz.block.ftl;

/**
 * A concrete definition of a topic. Extend this interface and annotate with {@code @TopicDefinition} to define a topic,
 * then inject this into verb methods to publish to the topic.
 *
 * @param <T>
 */
public interface Topic<T> {

    <T> void publish(T object);
}
