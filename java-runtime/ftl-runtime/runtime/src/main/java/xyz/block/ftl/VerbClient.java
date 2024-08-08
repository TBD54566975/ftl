package xyz.block.ftl;

import org.jetbrains.annotations.NotNull;

/**
 * A client for a specific verb
 * @param <P> The verb parameter type
 * @param <R> The verb return type
 */
public interface VerbClient<P, R> {

    R call(P param);

}
