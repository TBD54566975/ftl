package xyz.block.ftl;

/**
 * A client for a specific verb.
 *
 * The sink source and empty interfaces allow for different call signatures.
 *
 * @param <P> The verb parameter type
 * @param <R> The verb return type
 */
public interface VerbClient<P, R> {

    R call(P param);

}
