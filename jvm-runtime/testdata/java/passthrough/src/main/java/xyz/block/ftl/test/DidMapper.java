package xyz.block.ftl.test;

import java.nio.charset.StandardCharsets;

import ftl.gomodule.DidTypeAliasMapper;
import web5.sdk.dids.didcore.Did;

public class DidMapper implements DidTypeAliasMapper {

    @Override
    public Did decode(String bytes) {
        return Did.Parser.parse(bytes);
    }

    @Override
    public String encode(Did did) {
        return new String(did.marshalText(), StandardCharsets.UTF_8);
    }
}
