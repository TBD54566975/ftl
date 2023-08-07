package xyz.block.ftl.registry

import com.google.gson.Gson
import xyz.block.ftl.Context
import xyz.block.ftl.logging.Logging
import kotlin.reflect.KClass
import kotlin.reflect.KFunction
import kotlin.reflect.KParameter
import kotlin.reflect.full.createInstance

internal class VerbHandle<Resp>(
  private val verbClass: KClass<*>,
  private val verbFunction: KFunction<Resp>,
) {
  private val gson = Gson()

  private val logger = Logging.logger(VerbHandle::class)
  private val argumentType = findArgumentType(verbFunction.parameters)
  val returnType: KClass<*> = verbFunction.returnType.classifier as KClass<*>

  fun invokeVerbInternal(context: Context, argument: String): String {
    val instance = verbClass.createInstance()

    val arguments = verbFunction.parameters.associateWith { parameter ->
      when (parameter.type.classifier) {
        verbClass -> instance
        Context::class -> context
        else -> {
          val req = (parameter.type.classifier as KClass<*>).java
          gson.fromJson(argument, req)
        }
      }
    }

    val result = verbFunction.callBy(arguments)
    return gson.toJson(result)
  }

  private fun findArgumentType(parameters: List<KParameter>): KClass<*> {
    return parameters.find { param ->
      // skip the owning type itself
      null != param.name && Context::class != param.type.classifier
    }!!.type.classifier as KClass<*>
  }
}
