package xyz.block.ftl;

public interface TopicPartitionMapper<E> {
    String getPartitionKey(E event);
}