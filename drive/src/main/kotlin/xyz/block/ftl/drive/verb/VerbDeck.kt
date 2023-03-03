package xyz.block.ftl.drive.verb

import com.squareup.ftldemo.makePizza
import kotlin.reflect.KFunction1

class VerbDeck {
  data class VerbId(val qualifiedName: String)

  private val verbs = mapOf<VerbId, VerbCassette<out Any, out Any>>(
    toId(::makePizza) to VerbCassette(::makePizza)
  )

  private fun toId(verb: KFunction1<*, *>) = VerbId(verb.name)

  fun lookup(name: String): VerbCassette<out Any, out Any>? = verbs[VerbId(name)]
}
