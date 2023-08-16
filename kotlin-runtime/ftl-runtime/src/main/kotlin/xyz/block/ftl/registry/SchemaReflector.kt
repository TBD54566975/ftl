package xyz.block.ftl.registry

import xyz.block.ftl.Context
import xyz.block.ftl.Ingress
import xyz.block.ftl.v1.schema.*
import xyz.block.ftl.v1.schema.Array
import java.time.OffsetDateTime
import kotlin.Boolean
import kotlin.Int
import kotlin.Long
import kotlin.String
import kotlin.collections.Map
import kotlin.reflect.KClass
import kotlin.reflect.KFunction
import kotlin.reflect.KTypeParameter
import kotlin.reflect.full.createType
import kotlin.reflect.full.findAnnotation
import kotlin.reflect.full.hasAnnotation

internal fun reflectSchemaFromFunc(func: KFunction<*>): Module {
  if (!func.hasAnnotation<xyz.block.ftl.Verb>()) error("Function must be annotated with @Verb")
  if (func.parameters.size != 3) error("Verbs must have exactly two arguments")
  if (func.parameters[1].type.classifier != Context::class) error("First argument of verb must be Context")
  val requestType =
    func.parameters[2].type.classifier ?: error("Second argument of verb must be a data class")
  if (!(requestType as KClass<*>).isData) error("Second argument of verb must be a data class not $requestType")
  val returnType = func.returnType.classifier ?: error("Return type of verb must be a data class")
  if (!(returnType as KClass<*>).isData) error("Return type of verb must be a data class not $returnType")

  val requestData = reflectSchemaFromDataClass(requestType).map { Decl(data_ = it) }
  val responseData = reflectSchemaFromDataClass(returnType).map { Decl(data_ = it) }

  return Module(
    name = "",
    decls = listOf(
      Decl(verb = Verb(
        name = func.name,
        request = DataRef(name = requestType.simpleName!!),
        response = DataRef(name = returnType.simpleName!!),
        metadata = buildList {
          func.findAnnotation<Ingress>()?.let {
            add(Metadata(ingress = MetadataIngress(method = it.method.toString(), path = it.path)))
          }
        }
      ))
    ) + requestData + responseData
  )
}

internal fun reflectSchemaFromDataClass(dataClass: KClass<*>): List<Data> {
  if (!dataClass.isData) error("Must be a data class")
  return listOf(
    Data(
      name = dataClass.simpleName!!,
      fields = dataClass.constructors.first().parameters
        .filter { param -> param.name!! != "_empty" && param.type.classifier != Unit::class }
        .map { param ->
          Field(
            name = param.name!!,
            type = reflectType(param.type.classifier as KClass<*>),
          )
        }
    ))
}

internal fun reflectType(cls: KClass<*>): Type {
  return when (cls) {
    String::class -> Type(string = xyz.block.ftl.v1.schema.String())
    Int::class -> Type(int = xyz.block.ftl.v1.schema.Int())
    Long::class -> Type(int = xyz.block.ftl.v1.schema.Int())
    Boolean::class -> Type(bool = Bool())
    OffsetDateTime::class -> Type(time = Time())
    Map::class -> Type(
      map = xyz.block.ftl.v1.schema.Map(
        key = reflectTypeParameter(cls.typeParameters[0]),
        value_ = reflectTypeParameter(cls.typeParameters[1])
      )
    )

    List::class -> Type(
      array = Array(
        element = reflectTypeParameter(cls.typeParameters[0])
      )
    )

    else -> Type(dataRef = DataRef(name = cls.simpleName!!))
  }
}

internal fun reflectTypeParameter(param: KTypeParameter): Type {
  println("reflectTypeParameter: ${param.variance.name}")
  return when (param.createType()) {
    String::class -> Type(string = xyz.block.ftl.v1.schema.String())
    Int::class -> Type(int = xyz.block.ftl.v1.schema.Int())
    Long::class -> Type(int = xyz.block.ftl.v1.schema.Int())
    Boolean::class -> Type(bool = Bool())
    OffsetDateTime::class -> Type(time = Time())
    Map::class -> Type(
      map = xyz.block.ftl.v1.schema.Map(
        key = reflectType(param.upperBounds[0].classifier as KClass<*>),
        value_ = reflectType(param.upperBounds[1].classifier as KClass<*>)
      )
    )

    List::class -> Type(
      array = Array(
        element = reflectType(param.upperBounds[0].classifier as KClass<*>)
      )
    )

    else -> Type(dataRef = DataRef(name = param.name))
  }
}
