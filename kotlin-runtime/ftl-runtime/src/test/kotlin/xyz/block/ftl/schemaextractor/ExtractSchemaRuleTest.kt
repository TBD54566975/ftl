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
import xyz.block.ftl.v1.schema.*
import xyz.block.ftl.v1.schema.Array
import xyz.block.ftl.v1.schema.Map
import java.io.File
import kotlin.test.AfterTest

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
        package ftl.echo

        import ftl.builtin.Empty
        import ftl.builtin.HttpRequest
        import ftl.builtin.HttpResponse
        import ftl.time.TimeModuleClient
        import ftl.time.TimeRequest
        import ftl.time.TimeResponse
        import xyz.block.ftl.Alias
        import xyz.block.ftl.Context
        import xyz.block.ftl.HttpIngress
        import xyz.block.ftl.Method
        import xyz.block.ftl.Verb

        class InvalidInput(val field: String) : Exception()

        data class MapValue(val value: String)
        data class EchoMessage(val message: String, val metadata: Map<String, MapValue>? = null)

        /**
         * Request to echo a message.
         */
        data class EchoRequest<T>(val t: T, val name: String, @Alias("stf") val stuff: Any)
        data class EchoResponse(val messages: List<EchoMessage>)

        /**
         * Echo module.
         */
        class Echo {
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
                return context.call(TimeModuleClient::time, TimeRequest)
            }
        }
        """
    ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    val file = File(OUTPUT_FILENAME)
    val module = Module.ADAPTER.decode(file.inputStream())

    val expected = Module(
      name = "echo",
      comments = listOf(
        """/**
         * Echo module.
         */"""
      ),
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
                  map = Map(
                    key = Type(string = xyz.block.ftl.v1.schema.String()),
                    value_ = Type(
                      dataRef = DataRef(
                        name = "MapValue",
                        module = "echo"
                      )
                    )
                  )
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
                type = Type(dataRef = DataRef(name = "T"))
              ),
              Field(
                name = "name",
                type = Type(string = xyz.block.ftl.v1.schema.String())
              ),
              Field(
                name = "stuff",
                type = Type(any = xyz.block.ftl.v1.schema.Any()),
                alias = "stf"
              )
            ),
            comments = listOf(
              """/**
         * Request to echo a message.
         */"""
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
                      dataRef = DataRef(
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
              """/**
            * Echoes the given message.
            */"""
            ),
            request = Type(
              dataRef = DataRef(
                name = "HttpRequest",
                typeParameters = listOf(
                  Type(
                    dataRef = DataRef(
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
              dataRef = DataRef(
                name = "HttpResponse",
                module = "builtin",
                typeParameters = listOf(
                  Type(
                    dataRef = DataRef(
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
                    VerbRef(
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
              dataRef = DataRef(
                name = "Empty",
                module = "builtin"
              )
            ),
            response = Type(
              dataRef = DataRef(
                name = "Empty",
                module = "builtin"
              )
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
    val code = """
        package ftl.echo

        import ftl.time.TimeModuleClient
        import ftl.time.TimeRequest
        import ftl.time.TimeResponse
        import xyz.block.ftl.Context
        import xyz.block.ftl.HttpIngress
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
         * Echo module.
         */
        class Echo {
           /**
            * Echoes the given message.
            */
            @Throws(InvalidInput::class)
            @Verb
            @HttpIngress(Method.GET, "/echo")
            fun echo(context: Context, req: EchoRequest): EchoResponse {
                callTime(context)
                return EchoResponse(messages = listOf(EchoMessage(message = "Hello!")))
            }

            fun callTime(context: Context): TimeResponse {
                return context.call(TimeModuleClient::time, TimeRequest)
            }
        }
        """
    assertThrows<IllegalArgumentException>(message = "kotlin.Char type is not supported in FTL schema") {
      ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    }
  }

  @Test
  fun `fails if http ingress without http request-response types`() {
    val code = """
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
         * Echo module.
         */
        class Echo {
           /**
            * Echoes the given message.
            */
            @Throws(InvalidInput::class)
            @Verb
            @HttpIngress(Method.GET, "/echo")
            fun echo(context: Context, req: EchoRequest): EchoResponse {
                return EchoResponse(messages = listOf(EchoMessage(message = "Hello!")))
            }
        }
        """
    assertThrows<java.lang.IllegalArgumentException>(message = "@HttpIngress-annotated echo request must be ftl.builtin.HttpRequest") {
      ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    }
  }

  @AfterTest
  fun cleanup() {
    val file = File(OUTPUT_FILENAME)
    file.delete()
  }
}
