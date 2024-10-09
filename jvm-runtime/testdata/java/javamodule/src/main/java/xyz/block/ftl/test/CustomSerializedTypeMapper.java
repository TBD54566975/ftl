package xyz.block.ftl.test;

import xyz.block.ftl.TypeAlias;
import xyz.block.ftl.TypeAliasMapper;

@TypeAlias(name = "CustomSerializedType")
public class CustomSerializedTypeMapper implements TypeAliasMapper<CustomSerializedType, String> {
    @Override
    public String encode(CustomSerializedType object) {
        return object.getValue();
    }

    @Override
    public CustomSerializedType decode(String serialized) {
        return new CustomSerializedType(serialized);
    }
}
