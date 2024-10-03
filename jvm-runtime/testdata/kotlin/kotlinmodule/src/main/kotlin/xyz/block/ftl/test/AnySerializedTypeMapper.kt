package xyz.block.ftl.test

import com.fasterxml.jackson.databind.JsonNode
import com.fasterxml.jackson.databind.node.TextNode
import xyz.block.ftl.LanguageTypeMapping
import xyz.block.ftl.TypeAlias
import xyz.block.ftl.TypeAliasMapper

@TypeAlias(
  name = "AnySerializedType",
  languageTypeMappings = [LanguageTypeMapping(language = "go", type = "github.com/blockxyz/ftl/test.AnySerializedType")]
)
class AnySerializedTypeMapper : TypeAliasMapper<AnySerializedType, JsonNode> {
    override fun encode(`object`: AnySerializedType): JsonNode {
        return TextNode.valueOf(`object`.value)
    }

    override fun decode(serialized: JsonNode): AnySerializedType {
        if (serialized.isTextual) {
            return AnySerializedType(serialized.textValue())
        }
        throw RuntimeException("Expected a textual value")
    }
}
