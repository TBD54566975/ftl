package xyz.block.ftl.registry

import xyz.block.ftl.Context
import xyz.block.ftl.Ingress
import xyz.block.ftl.v1.schema.DataRef
import xyz.block.ftl.v1.schema.Metadata
import xyz.block.ftl.v1.schema.MetadataIngress
import xyz.block.ftl.v1.schema.Verb
import kotlin.reflect.KClass
import kotlin.reflect.KFunction
import kotlin.reflect.full.findAnnotation
import kotlin.reflect.full.hasAnnotation

internal fun reflectSchemaFromFunc(func: KFunction<*>): Verb? {
  if (!func.hasAnnotation<xyz.block.ftl.Verb>()) return null
  if (func.parameters.size != 3) error("Verbs must have exactly two arguments")
  if (func.parameters[1].type.classifier != Context::class) error("First argument of verb must be Context")
  val requestType =
    func.parameters[2].type.classifier ?: error("Second argument of verb must be a data class")
  if (!(requestType as KClass<*>).isData) error("Second argument of verb must be a data class not $requestType")
  val returnType = func.returnType.classifier ?: error("Return type of verb must be a data class")
  if (!(returnType as KClass<*>).isData) error("Return type of verb must be a data class not $returnType")

  return Verb(
    name = func.name,
    request = DataRef(name = requestType.simpleName!!),
    response = DataRef(name = returnType.simpleName!!),
    metadata = buildList {
      func.findAnnotation<Ingress>()?.let {
        add(Metadata(ingress = MetadataIngress(method = it.method.toString(), path = it.path)))
      }
    }
  )
}