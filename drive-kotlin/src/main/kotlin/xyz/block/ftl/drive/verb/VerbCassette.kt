package xyz.block.ftl.drive.verb

import xyz.block.ftl.Context
import kotlin.reflect.KClass
import kotlin.reflect.KFunction
import kotlin.reflect.KParameter
import kotlin.reflect.full.isSubtypeOf
import kotlin.reflect.full.staticProperties

class VerbCassette<R>(val verbId: VerbDeck.VerbId, private val verb: KFunction<R>) {
  val argumentType = findArgumentType(verb.parameters)
  val returnType: KClass<*> = verb.returnType.classifier as KClass<*>

  fun dispatch(context: Context, argument: Any): R {
    val arguments = verb.parameters.associateWith { parameter ->
      when (parameter.type.classifier) {
        Context::class -> context
        else -> argument
      }
    }

    return verb.callBy(arguments)
  }

  private fun findArgumentType(parameters: List<KParameter>): KClass<*> {
    return parameters.find {
        param -> Context::class != param.type.classifier
    }!!.type.classifier as KClass<*>
  }
}
