package xyz.block.ftl.drive.verb

import xyz.block.ftl.Context
import xyz.block.ftl.SimpleVerb
import xyz.block.ftl.drive.Logging
import kotlin.reflect.KClass
import kotlin.reflect.KFunction
import kotlin.reflect.KParameter
import kotlin.reflect.full.createInstance

class VerbCassette<R>(
  val verbId: VerbDeck.VerbId,
  private val verbClass: KClass<*>,
  private val verbFunction: KFunction<R>) {

  private val logger = Logging.logger(VerbCassette::class)
  private val argumentType = findArgumentType(verbFunction.parameters)
  val returnType: KClass<*> = verbFunction.returnType.classifier as KClass<*>

  fun invokeVerbInternal(context: Context, argument: Any): R {
    val instance = verbClass.createInstance()

    val arguments = verbFunction.parameters.associateWith { parameter ->
      when (parameter.type.classifier) {
        verbClass -> instance
        Context::class -> context
        else -> argument
      }
    }

    return verbFunction.callBy(arguments)
  }

  private fun findArgumentType(parameters: List<KParameter>): KClass<*> {
    return parameters.find { param ->
      // skip the owning type itself
      null != param.name
        &&
      Context::class != param.type.classifier
    }!!.type.classifier as KClass<*>
  }

  fun readSchema() {
    val schema = (verbClass.createInstance() as SimpleVerb).schema()

    logger.debug("    processing: ${verbId.qualifiedName} (${schema})")
  }

  fun toDescriptor(): VerbDeck.VerbSignature = VerbDeck.VerbSignature(verbId, argumentType)
}
