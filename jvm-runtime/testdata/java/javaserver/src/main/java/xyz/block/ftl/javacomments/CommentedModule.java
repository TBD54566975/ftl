package xyz.block.ftl.javacomments;

import org.jetbrains.annotations.NotNull;

import xyz.block.ftl.Config;
import xyz.block.ftl.Export;
import xyz.block.ftl.Secret;
import xyz.block.ftl.Verb;

/**
 * Comment on a module class
 */
public class CommentedModule {

    /**
     * Comment on a verb
     *
     * @param val Parameter comment
     * @param configString Config comment
     * @param secretString Secret comment
     * @return Great success
     */
    @Export
    @Verb
    public @NotNull EnumType MultilineCommentVerb(
            @NotNull DataClass val,
            @Config("config") String configString,
            @Secret("secretString") String secretString) {
        return EnumType.PORTENTOUS;
    }

    //TODO TypeAlias, Database, Topic, Subscription, Lease, Cron
}
