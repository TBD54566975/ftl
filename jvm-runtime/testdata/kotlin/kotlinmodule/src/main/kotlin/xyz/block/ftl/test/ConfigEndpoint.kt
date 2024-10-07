package xyz.block.ftl.test

import xyz.block.ftl.Config
import xyz.block.ftl.Verb

class ConfigEndpoint {
    @Verb
    fun config(@Config("key") key: String): String {
        return key
    }
}
