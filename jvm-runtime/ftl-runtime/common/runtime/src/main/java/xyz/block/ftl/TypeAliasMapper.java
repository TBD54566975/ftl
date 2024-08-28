package xyz.block.ftl;

public interface TypeAliasMapper<T, S> {

    S encode(T object);

    T decode(S serialized);

}
