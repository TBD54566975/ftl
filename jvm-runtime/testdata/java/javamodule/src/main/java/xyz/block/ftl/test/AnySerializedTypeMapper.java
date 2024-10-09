package xyz.block.ftl.test;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.TextNode;

import xyz.block.ftl.LanguageTypeMapping;
import xyz.block.ftl.TypeAlias;
import xyz.block.ftl.TypeAliasMapper;

@TypeAlias(name = "AnySerializedType", languageTypeMappings = {
        @LanguageTypeMapping(language = "go", type = "github.com/blockxyz/ftl/test.AnySerializedType"),
})
public class AnySerializedTypeMapper implements TypeAliasMapper<AnySerializedType, JsonNode> {
    @Override
    public JsonNode encode(AnySerializedType object) {
        return TextNode.valueOf(object.getValue());
    }

    @Override
    public AnySerializedType decode(JsonNode serialized) {
        if (serialized.isTextual()) {
            return new AnySerializedType(serialized.textValue());
        }
        throw new RuntimeException("Expected a textual value");
    }
}
