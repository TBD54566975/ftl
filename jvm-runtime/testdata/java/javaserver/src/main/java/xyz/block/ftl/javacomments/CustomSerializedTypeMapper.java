package xyz.block.ftl.test;

import xyz.block.ftl.TypeAlias;
import xyz.block.ftl.TypeAliasMapper;

/**
 * Comment on a TypeAlias
 */
@TypeAlias(name = "CustomSerializedType")
public class CustomSerializedTypeMapper implements TypeAliasMapper<xyz.block.ftl.test.CustomSerializedType, String> {
    @Override
    public String encode(xyz.block.ftl.test.CustomSerializedType object) {
        return object.getValue();
    }

    @Override
    public xyz.block.ftl.test.CustomSerializedType decode(String serialized) {
        return new xyz.block.ftl.test.CustomSerializedType(serialized);
    }
}
