package xyz.block.ftl.deployment;

import java.util.Collection;
import java.util.List;
import java.util.Map;

import io.quarkus.builder.item.SimpleBuildItem;

public final class CommentsBuildItem extends SimpleBuildItem {

    final Map<String, Collection<String>> comments;

    public CommentsBuildItem(Map<String, Collection<String>> comments) {
        this.comments = comments;
    }

    public Iterable<String> getComments(String name) {
        return comments.getOrDefault(name, List.of());
    }
}
