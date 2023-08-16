package xyz.block.ftl.registry

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import xyz.block.ftl.Context
import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.v1.schema.*
import xyz.block.ftl.v1.schema.String as SchemaString

data class Value(val value: String)

data class SchemaExampleVerbRequest(
  val string: String,
  val int: Int,
  val long: Long,
  val map: Map<String, Value>
)

data class SchemaExampleVerbResponse(
  val _empty: Unit = Unit,
)

class SchemaExampleVerb {
  @xyz.block.ftl.Verb
  @Ingress(Method.GET, "/test")
  fun verb(context: Context, req: SchemaExampleVerbRequest): SchemaExampleVerbResponse {
    return SchemaExampleVerbResponse()
  }
}

class SchemaReflectorKtTest {
  @Test
  fun reflectSchemaFromFunc() {
    val expected = Module(
      decls = listOf(
        Decl(
          verb = Verb(
            name = "verb",
            request = DataRef(name = "VerbRequest"),
            response = DataRef(name = "VerbResponse"),
            metadata = listOf(
              Metadata(ingress = MetadataIngress(method = "GET", path = "/test")),
            ),
          )
        ),
        Decl(
          data_ = Data(
            name = "VerbRequest",
            fields = listOf(
              Field(name = "text", type = Type(string = SchemaString())),
            )
          )
        ),
        Decl(
          data_ = Data(
            name = "VerbResponse",
            fields = listOf(
              Field(name = "text", type = Type(string = SchemaString())),
            )
          )
        ),
      )
    )
    val actual = reflectSchemaFromFunc(SchemaExampleVerb::verb)
    println(actual)
    assertEquals(expected, actual)
  }
}
