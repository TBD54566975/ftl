package xyz.block.ftl.test

import ftl.gomodule.DidTypeAliasMapper
import web5.sdk.dids.didcore.Did


class DidMapper : DidTypeAliasMapper
{

  override fun decode(bytes: String): Did {
    return Did.Parser.parse(bytes)
  }

  override fun encode(did: Did): String {
    return String(did.marshalText())
  }

}
