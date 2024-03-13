package xyz.block.ftl

import xyz.block.ftl.client.VerbServiceClient
import xyz.block.ftl.registry.Ref
import xyz.block.ftl.registry.ftlModuleFromJvmModule
import xyz.block.ftl.serializer.makeGson
import java.security.InvalidParameterException
import kotlin.jvm.internal.CallableReference
import kotlin.reflect.KFunction
import kotlin.reflect.full.hasAnnotation

class Context(
  val jvmModule: String,
  val routingClient: VerbServiceClient,
) {
  val gson = makeGson()

  /// Class method with Context.
  inline fun <reified R> call(verb: KFunction<R>, request: Any): R {
    if (!verb.hasAnnotation<Verb>()) throw InvalidParameterException(
      "verb must be annotated with @Verb"
    )
    if (verb !is CallableReference) {
      throw InvalidParameterException("could not determine module from verb name")
    }
    val ftlModule = ftlModuleFromJvmModule(jvmModule, verb)
    val requestJson = gson.toJson(request)
    val responseJson = routingClient.call(this, Ref(ftlModule, verb.name), requestJson)
    return gson.fromJson(responseJson, R::class.java)
  }
}
