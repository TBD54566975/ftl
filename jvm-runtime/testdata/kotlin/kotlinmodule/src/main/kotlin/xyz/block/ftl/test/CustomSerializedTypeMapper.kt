package xyz.block.ftl.test

import xyz.block.ftl.TypeAlias
import xyz.block.ftl.TypeAliasMapper

@TypeAlias(name = "CustomSerializedType")
class CustomSerializedTypeMapper : TypeAliasMapper<CustomSerializedType, String> {
    override fun encode(`object`: CustomSerializedType): String {
        return `object`.value
    }

    override fun decode(serialized: String): CustomSerializedType {
        return CustomSerializedType(serialized)
    }
}
