package xyz.block.ftl.drive.verb

import kotlin.reflect.KFunction1

class VerbCassette<A, R>(private val verb: KFunction1<A, R>) {

  fun invokeVerb(argument: A): String {
    return verb.invoke(argument).toString()
  }
}
