package xyz.block.ftl.java.runtime.it;

import xyz.block.ftl.Export;
import xyz.block.ftl.Topic;
import xyz.block.ftl.TopicDefinition;

@Export
@TopicDefinition(value = "testTopic")
public interface MyTopic extends Topic<Person> {
}
