package xyz.block.ftl.registry

import xyz.block.ftl.Context
import xyz.block.ftl.serializer.makeGson
import kotlin.reflect.KClass
import kotlin.reflect.KFunction
import kotlin.reflect.full.createInstance
import kotlin.reflect.jvm.javaType

internal class VerbHandle<Resp>(
  private val verbClass: KClass<*>,
  private val verbFunction: KFunction<Resp>,
) {
  private val gson = makeGson()

  fun invokeVerbInternal(context: Context, argument: String): String {
    val instance = verbClass.createInstance()

    val arguments = verbFunction.parameters.associateWith { parameter ->
      when (parameter.type.classifier) {
        verbClass -> instance
        Context::class -> context
        else -> {
          gson.fromJson(argument, parameter.type.javaType)
        }
      }
    }

    val result = verbFunction.callBy(arguments)
    return gson.toJson(result)
  }
}
