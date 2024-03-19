package xyz.block.ftl.schemaextractor

import io.gitlab.arturbosch.detekt.api.Config
import io.gitlab.arturbosch.detekt.rules.KotlinCoreEnvironmentTest
import io.gitlab.arturbosch.detekt.test.compileAndLintWithContext
import org.assertj.core.api.Assertions.assertThat
import org.jetbrains.kotlin.cli.jvm.compiler.KotlinCoreEnvironment
import org.jetbrains.kotlin.cli.jvm.config.addJvmClasspathRoots
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import xyz.block.ftl.schemaextractor.ExtractSchemaRule.Companion.OUTPUT_FILENAME
import xyz.block.ftl.v1.schema.Array
import xyz.block.ftl.v1.schema.Data
import xyz.block.ftl.v1.schema.Decl
import xyz.block.ftl.v1.schema.Enum
import xyz.block.ftl.v1.schema.EnumVariant
import xyz.block.ftl.v1.schema.Field
import xyz.block.ftl.v1.schema.IngressPathComponent
import xyz.block.ftl.v1.schema.IngressPathLiteral
import xyz.block.ftl.v1.schema.IntValue
import xyz.block.ftl.v1.schema.Map
import xyz.block.ftl.v1.schema.Metadata
import xyz.block.ftl.v1.schema.MetadataAlias
import xyz.block.ftl.v1.schema.MetadataCalls
import xyz.block.ftl.v1.schema.MetadataIngress
import xyz.block.ftl.v1.schema.Module
import xyz.block.ftl.v1.schema.Optional
import xyz.block.ftl.v1.schema.Position
import xyz.block.ftl.v1.schema.Ref
import xyz.block.ftl.v1.schema.String
import xyz.block.ftl.v1.schema.StringValue
import xyz.block.ftl.v1.schema.Type
import xyz.block.ftl.v1.schema.TypeParameter
import xyz.block.ftl.v1.schema.Unit
import xyz.block.ftl.v1.schema.Value
import xyz.block.ftl.v1.schema.Verb
import java.io.File
import kotlin.test.AfterTest
import kotlin.test.assertContains

@KotlinCoreEnvironmentTest
internal class ExtractSchemaRuleTest(private val env: KotlinCoreEnvironment) {
  @BeforeEach
  fun setup() {
    val dependenciesDir = File("src/test/kotlin/xyz/block/ftl/schemaextractor/testdata/dependencies")
    val dependencies = dependenciesDir.listFiles { file -> file.extension == "kt" }?.toList() ?: emptyList()
    env.configuration.addJvmClasspathRoots(dependencies)
  }

  @Test
  fun `extracts schema`() {
    val code = """
      // Echo module.
      package ftl.echo

      import ftl.builtin.Empty
      import ftl.builtin.HttpRequest
      import ftl.builtin.HttpResponse
      import ftl.time.time as verb
      import ftl.time.other
      import ftl.time.TimeRequest
      import ftl.time.TimeResponse
      import xyz.block.ftl.Json
      import xyz.block.ftl.Context
      import xyz.block.ftl.HttpIngress
      import xyz.block.ftl.Method
      import xyz.block.ftl.Module
      import xyz.block.ftl.Verb

      class InvalidInput(val field: String) : Exception()

      data class MapValue(val value: String)
      data class EchoMessage(val message: String, val metadata: Map<String, MapValue>? = null)

      /**
       * Request to echo a message.
       *
       * More comments.
       */
      data class EchoRequest<T>(
        val t: T,
        val name: String,
        @Json("stf") val stuff: Any,
       )
      data class EchoResponse(val messages: List<EchoMessage>)

      /**
       * Echoes the given message.
       */
      @Throws(InvalidInput::class)
      @Verb
      @HttpIngress(Method.GET, "/echo")
      fun echo(context: Context, req: HttpRequest<EchoRequest<String>>): HttpResponse<EchoResponse, String> {
        callTime(context)

        return HttpResponse(
          status = 200,
          headers = mapOf("Get" to arrayListOf("Header from FTL")),
          body = EchoResponse(messages = listOf(EchoMessage(message = "Hello!")))
        )
      }

      @Verb
      fun empty(context: Context, req: Empty): Empty {
        return builtin.Empty()
      }

      fun callTime(context: Context): TimeResponse {
        context.call(::empty, builtin.Empty())
        context.call(::other, builtin.Empty())
        // commented out call is ignored:
        //context.call(::foo, builtin.Empty())
        return context.call(::verb, builtin.Empty())
      }

      @Verb
      fun sink(context: Context, req: Empty) {}

      @Verb
      fun source(context: Context): Empty {}

      @Verb
      fun emptyVerb(context: Context) {}
    """
    ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    val file = File(OUTPUT_FILENAME)
    val module = Module.ADAPTER.decode(file.inputStream())

    val expected = Module(
      name = "echo",
      comments = listOf("Echo module."),
      decls = listOf(
        Decl(
          data_ = Data(
            name = "MapValue",
            fields = listOf(
              Field(
                name = "value",
                type = Type(string = xyz.block.ftl.v1.schema.String())
              )
            ),
          ),
        ),
        Decl(
          data_ = Data(
            name = "EchoMessage",
            fields = listOf(
              Field(
                name = "message",
                type = Type(string = xyz.block.ftl.v1.schema.String())
              ),
              Field(
                name = "metadata",
                type = Type(
                  optional = Optional(
                    type = Type(
                      map = Map(
                        key = Type(string = xyz.block.ftl.v1.schema.String()),
                        value_ = Type(
                          ref = Ref(
                            name = "MapValue",
                            module = "echo"
                          )
                        )
                      )
                    )
                  ),
                )
              ),
            ),
          ),
        ),
        Decl(
          data_ = Data(
            name = "EchoRequest",
            fields = listOf(
              Field(
                name = "t",
                type = Type(ref = Ref(name = "T"))
              ),
              Field(
                name = "name",
                type = Type(string = xyz.block.ftl.v1.schema.String())
              ),
              Field(
                name = "stuff",
                type = Type(any = xyz.block.ftl.v1.schema.Any()),
                metadata = listOf(Metadata(alias = MetadataAlias(alias = "stf"))),
              )
            ),
            comments = listOf(
              "Request to echo a message.", "", "More comments."
            ),
            typeParameters = listOf(
              TypeParameter(name = "T")
            )
          ),
        ),
        Decl(
          data_ = Data(
            name = "EchoResponse",
            fields = listOf(
              Field(
                name = "messages",
                type = Type(
                  array = Array(
                    element = Type(
                      ref = Ref(
                        name = "EchoMessage",
                        module = "echo"
                      )
                    )
                  )
                )
              )
            ),
          ),
        ),
        Decl(
          verb = Verb(
            name = "echo",
            comments = listOf(
              """Echoes the given message."""
            ),
            request = Type(
              ref = Ref(
                name = "HttpRequest",
                typeParameters = listOf(
                  Type(
                    ref = Ref(
                      name = "EchoRequest",
                      typeParameters = listOf(
                        Type(string = xyz.block.ftl.v1.schema.String())
                      ),
                      module = "echo"
                    )
                  )
                ),
                module = "builtin"
              )
            ),
            response = Type(
              ref = Ref(
                name = "HttpResponse",
                module = "builtin",
                typeParameters = listOf(
                  Type(
                    ref = Ref(
                      name = "EchoResponse",
                      module = "echo"
                    )
                  ),
                  Type(
                    string = xyz.block.ftl.v1.schema.String()
                  ),
                ),
              ),
            ),
            metadata = listOf(
              Metadata(
                ingress = MetadataIngress(
                  type = "http",
                  method = "GET",
                  path = listOf(
                    IngressPathComponent(
                      ingressPathLiteral = IngressPathLiteral(text = "echo")
                    )
                  )
                )
              ),
              Metadata(
                calls = MetadataCalls(
                  calls = listOf(
                    Ref(
                      name = "empty",
                      module = "echo"
                    ),
                    Ref(
                      name = "other",
                      module = "time"
                    ),
                    Ref(
                      name = "time",
                      module = "time"
                    )
                  )
                )
              )
            )
          ),
        ),
        Decl(
          verb = Verb(
            name = "empty",
            request = Type(
              ref = Ref(
                name = "Empty",
                module = "builtin"
              )
            ),
            response = Type(
              ref = Ref(
                name = "Empty",
                module = "builtin"
              )
            ),
          ),
        ),
        Decl(
          verb = Verb(
            name = "sink",
            request = Type(
              ref = Ref(
                name = "Empty",
                module = "builtin"
              )
            ),
            response = Type(
              unit = Unit()
            ),
          ),
        ),
        Decl(
          verb = Verb(
            name = "source",
            request = Type(
              unit = Unit()
            ),
            response = Type(
              ref = Ref(
                name = "Empty",
                module = "builtin"
              )
            ),
          ),
        ),
        Decl(
          verb = Verb(
            name = "emptyVerb",
            request = Type(
              unit = Unit()
            ),
            response = Type(
              unit = Unit()
            ),
          ),
        )
      )
    )

    assertThat(module)
      .usingRecursiveComparison()
      .withEqualsForType({ _, _ -> true }, Position::class.java)
      .ignoringFieldsMatchingRegexes(".*hashCode\$")
      .isEqualTo(expected)
  }

  @Test
  fun `fails if invalid schema type is included`() {
    val code = """/**
 * Echo module.
 */
package ftl.echo

import ftl.builtin.Empty
import ftl.time.time
import ftl.time.TimeRequest
import ftl.time.TimeResponse
import xyz.block.ftl.Context
import xyz.block.ftl.Method
import xyz.block.ftl.Verb

class InvalidInput(val field: String) : Exception()

data class EchoMessage(val message: String, val metadata: Map<String, String>? = null)

/**
 * Request to echo a message.
 */
data class EchoRequest(val name: Char)
data class EchoResponse(val messages: List<EchoMessage>)

/**
 * Echoes the given message.
 */
@Throws(InvalidInput::class)
@Verb
fun echo(context: Context, req: EchoRequest): EchoResponse {
  callTime(context)
  return EchoResponse(messages = listOf(EchoMessage(message = "Hello!")))
}

fun callTime(context: Context): TimeResponse {
  return context.call(::time, Empty())
}
"""
    val message = assertThrows<IllegalArgumentException> {
      ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    }.message!!
    assertContains(message, "Expected type to be a data class or builtin.Empty, but was kotlin.Char")
  }

  @Test
  fun `fails if http ingress without http request-response types`() {
    val code = """
 /**
 * Echo module.
 */
package ftl.echo

import xyz.block.ftl.Context
import xyz.block.ftl.HttpIngress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb

/**
 * Request to echo a message.
 */
data class EchoRequest(val name: String)
data class EchoResponse(val message: String)

/**
 * Echoes the given message.
 */
@Throws(InvalidInput::class)
@Verb
@HttpIngress(Method.GET, "/echo")
fun echo(context: Context, req: EchoRequest): EchoResponse {
  return EchoResponse(messages = listOf(EchoMessage(message = "Hello!")))
}
        """
    val message = assertThrows<java.lang.IllegalArgumentException> {
      ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    }.message!!
    assertContains(message, "@HttpIngress-annotated echo request must be ftl.builtin.HttpRequest")
  }

  @Test
  fun `source and sink types`() {
    val code = """
 /**
 * Echo module.
 */
package ftl.echo

import xyz.block.ftl.Context
import xyz.block.ftl.HttpIngress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb

/**
 * Request to echo a message.
 */
data class EchoRequest(val name: String)
data class EchoResponse(val message: String)

/**
 * Echoes the given message.
 */
@Throws(InvalidInput::class)
@Verb
@HttpIngress(Method.GET, "/echo")
fun echo(context: Context, req: EchoRequest): EchoResponse {
  return EchoResponse(messages = listOf(EchoMessage(message = "Hello!")))
}
        """
    val message = assertThrows<java.lang.IllegalArgumentException> {
      ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    }.message!!
    assertContains(message, "@HttpIngress-annotated echo request must be ftl.builtin.HttpRequest")
  }

  @AfterTest
  fun cleanup() {
    val file = File(OUTPUT_FILENAME)
    file.delete()
  }

  @Test
  fun `extracts enums`() {
    val code = """
      package ftl.things

      import ftl.time.Color
      import xyz.block.ftl.Json
      import xyz.block.ftl.Context
      import xyz.block.ftl.Method
      import xyz.block.ftl.Verb

      class InvalidInput(val field: String) : Exception()

      enum class Thing {
       /**
        * A comment.
        */
        A,
        B,
        C,
      }

      /**
       * Comments.
       */
      enum class StringThing(val value: String) {
        /**
         * A comment.
         */
        A("A"),
        /**
         * B comment.
         */
        B("B"),
        C("C"),
      }

      enum class IntThing(val value: Int) {
        A(1),
        B(2),
        /**
         * C comment.
         */
        C(3),
      }

      data class Request(
        val color: Color,
        val thing: Thing,
        val stringThing: StringThing,
        val intThing: IntThing
      )

      data class Response(val message: String)

      @Verb
      fun something(context: Context, req: Request): Response {
        return Response(message = "response")
      }
    """
    ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    val file = File(OUTPUT_FILENAME)
    val module = Module.ADAPTER.decode(file.inputStream())

    val expected = Module(
      name = "things",
      decls = listOf(
        Decl(
          data_ = Data(
            name = "Request",
            fields = listOf(
              Field(
                name = "color",
                type = Type(ref = Ref(name = "Color", module = "time"))
              ),
              Field(
                name = "thing",
                type = Type(ref = Ref(name = "Thing", module = "things"))
              ),
              Field(
                name = "stringThing",
                type = Type(ref = Ref(name = "StringThing", module = "things"))
              ),
              Field(
                name = "intThing",
                type = Type(ref = Ref(name = "IntThing", module = "things"))
              ),
            ),
          ),
        ),
        Decl(
          data_ = Data(
            name = "Response",
            fields = listOf(
              Field(
                name = "message",
                type = Type(string = xyz.block.ftl.v1.schema.String())
              )
            ),
          ),
        ),
        Decl(
          verb = Verb(
            name = "something",
            request = Type(ref = Ref(name = "Request", module = "things")),
            response = Type(ref = Ref(name = "Response", module = "things")),
          ),
        ),
        Decl(
          enum_ = Enum(
            name = "Thing",
            variants = listOf(
              EnumVariant(name = "A", value_ = Value(intValue = IntValue(value_ = 0)), comments = listOf("A comment.")),
              EnumVariant(name = "B", value_ = Value(intValue = IntValue(value_ = 1))),
              EnumVariant(name = "C", value_ = Value(intValue = IntValue(value_ = 2))),
            ),
          ),
        ),
        Decl(
          enum_ = Enum(
            name = "StringThing",
            comments = listOf("Comments."),
            variants = listOf(
              EnumVariant(
                name = "A",
                value_ = Value(stringValue = StringValue(value_ = "A")),
                comments = listOf("A comment.")
              ),
              EnumVariant(
                name = "B",
                value_ = Value(stringValue = StringValue(value_ = "B")),
                comments = listOf("B comment.")
              ),
              EnumVariant(name = "C", value_ = Value(stringValue = StringValue(value_ = "C"))),
            ),
          ),
        ),
        Decl(
          enum_ = Enum(
            name = "IntThing",
            variants = listOf(
              EnumVariant(name = "A", value_ = Value(intValue = IntValue(value_ = 1))),
              EnumVariant(name = "B", value_ = Value(intValue = IntValue(value_ = 2))),
              EnumVariant(name = "C", value_ = Value(intValue = IntValue(value_ = 3)), comments = listOf("C comment.")),
            ),
          ),
        ),
      )
    )

    assertThat(module)
      .usingRecursiveComparison()
      .withEqualsForType({ _, _ -> true }, Position::class.java)
      .ignoringFieldsMatchingRegexes(".*hashCode\$")
      .isEqualTo(expected)
  }


  @Test
  fun `extracts secrets and configs`() {
    val code = """
      package ftl.test

      import ftl.time.Color
      import xyz.block.ftl.Json
      import xyz.block.ftl.Context
      import xyz.block.ftl.Method
      import xyz.block.ftl.Verb
      import xyz.block.ftl.config.Config
      import xyz.block.ftl.secrets.Secret

      val secret = Secret.new<String>("secret")
      val anotherSecret = Secret(String::class.java, "anotherSecret")

      val config = Config.new<ConfigData>("config")
      val anotherConfig = Config(ConfigData::class.java, "anotherConfig")

      data class ConfigData(val field: String)

      data class Request(val message: String)

      data class Response(val message: String)

      @Verb
      fun something(context: Context, req: Request): Response {
        return Response(message = "response")
      }
    """
    ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    val file = File(OUTPUT_FILENAME)
    val module = Module.ADAPTER.decode(file.inputStream())

    val expected = Module(
      name = "test",
      decls = listOf(
        Decl(
          data_ = Data(
            name = "ConfigData",
            fields = listOf(
              Field(
                name = "field",
                type = Type(string = String())
              )
            ),
          ),
        ),
        Decl(
          data_ = Data(
            name = "Request",
            fields = listOf(
              Field(
                name = "message",
                type = Type(string = String())
              ),
            ),
          ),
        ),
        Decl(
          data_ = Data(
            name = "Response",
            fields = listOf(
              Field(
                name = "message",
                type = Type(string = String())
              )
            ),
          ),
        ),
        Decl(
          verb = Verb(
            name = "something",
            request = Type(ref = Ref(name = "Request", module = "test")),
            response = Type(ref = Ref(name = "Response", module = "test")),
          ),
        ),
        Decl(
          config = xyz.block.ftl.v1.schema.Config(
            name = "config",
            type = Type(ref = Ref(name = "ConfigData", module = "test")),
          ),
        ),
        Decl(
          config = xyz.block.ftl.v1.schema.Config(
            name = "anotherConfig",
            type = Type(ref = Ref(name = "ConfigData", module = "test")),
          ),
        ),
        Decl(
          secret = xyz.block.ftl.v1.schema.Secret(
            name = "secret",
            type = Type(string = String())
          ),
        ),
        Decl(
          secret = xyz.block.ftl.v1.schema.Secret(
            name = "anotherSecret",
            type = Type(string = String())
          ),
        ),
      )
    )

    assertThat(module)
      .usingRecursiveComparison()
      .withEqualsForType({ _, _ -> true }, Position::class.java)
      .ignoringFieldsMatchingRegexes(".*hashCode\$")
      .isEqualTo(expected)
  }
}
