package xyz.block.ftl.registry

import xyz.block.ftl.Context
import xyz.block.ftl.serializer.makeGson
import kotlin.reflect.KFunction
import kotlin.reflect.jvm.javaType

internal class VerbHandle<Resp>(
  private val verbFunction: KFunction<Resp>,
) {
  private val gson = makeGson()

  fun invokeVerbInternal(context: Context, argument: String): String {
    val arguments = verbFunction.parameters.associateWith { parameter ->
      when (parameter.type.classifier) {
        Context::class -> context
        else -> {
          val deserialized: Any? = gson.fromJson(argument, parameter.type.javaType)
          return@associateWith deserialized
        }
      }
    }

    val result = verbFunction.callBy(arguments)
    return gson.toJson(result)
  }
}
