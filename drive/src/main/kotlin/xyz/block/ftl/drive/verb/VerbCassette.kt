package xyz.block.ftl.drive.verb

import kotlin.reflect.KClass
import kotlin.reflect.KFunction1

@Suppress("UNCHECKED_CAST")
class VerbCassette<A, R>(private val verb: KFunction1<A, R>) {
  val argumentType = verb.parameters[0].type.classifier as KClass<*>
  val returnType: KClass<*> = verb.returnType.classifier as KClass<*>

  fun invokeVerb(argument: Any): Any {
    return verb.invoke(argument as A) as Any
  }
}
