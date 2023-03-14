package xyz.block.ftl

import xyz.block.ftl.drive.verb.VerbDeck
import kotlin.reflect.KFunction

class Ftl {
  companion object {
    fun <R> call(verb: KFunction<R>, request: Any): R {
      @Suppress("UNCHECKED_CAST")
      return VerbDeck.instance.dispatch(verb, request) as R
    }
  }
}
