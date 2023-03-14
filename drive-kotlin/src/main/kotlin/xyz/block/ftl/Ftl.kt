package xyz.block.ftl

import xyz.block.ftl.drive.verb.VerbDeck
import kotlin.reflect.KFunction

class Ftl {
  companion object {
    fun <R> call(context: Context, verb: KFunction<R>, request: Any): R {
      @Suppress("UNCHECKED_CAST")
      return VerbDeck.instance.dispatch(context, verb, request) as R
    }
  }
}
