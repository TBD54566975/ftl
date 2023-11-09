package xyz.block.ftl.schemaextractor

import io.gitlab.arturbosch.detekt.api.Config
import io.gitlab.arturbosch.detekt.test.compileAndLintWithContext
import org.jetbrains.kotlin.cli.jvm.compiler.KotlinCoreEnvironment
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import xyz.block.ftl.schemaextractor.ExtractSchemaRule.Companion.OUTPUT_FILENAME
import xyz.block.ftl.v1.schema.*
import xyz.block.ftl.v1.schema.Array
import xyz.block.ftl.v1.schema.Map
import java.io.File
import kotlin.test.AfterTest
import kotlin.test.Ignore


// TODO(@worstell):
// This test can't run when org.jetbrains.kotlin:kotlin-main-kts is excluded from the
// io.gitlab.arturbosch.detekt:detekt-test dependency, which is necessary to avoid a dependency conflict with
// Logback/SLF4J. When fixed we should uncomment the @KotlinCoreEnvironmentTest annotation amd no longer ignore this
// test.
@Ignore
//@KotlinCoreEnvironmentTest
internal class ExtractSchemaRuleTest(private val env: KotlinCoreEnvironment) {

  @Test
  fun `extracts schema`() {
    val code = """
        package ftl.echo

        import ftl.time.TimeModuleClient
        import ftl.time.TimeRequest
        import ftl.time.TimeResponse
        import xyz.block.ftl.Context
        import xyz.block.ftl.Ingress
        import xyz.block.ftl.Method
        import xyz.block.ftl.Verb

        class InvalidInput(val field: String) : Exception()

        data class MapValue(val value: String)
        data class EchoMessage(val message: String, val metadata: Map<String, MapValue>? = null)

        /**
         * Request to echo a message.
         */
        data class EchoRequest(val name: String)
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
            @Ingress(Method.GET, "/echo")
            fun echo(context: Context, req: EchoRequest): EchoResponse {
                callTime(context)
                return EchoResponse(messages = listOf(EchoMessage(message = "Hello!")))
            }

            fun callTime(context: Context): TimeResponse {
                return context.call(TimeModuleClient::time, TimeRequest())
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
            name = "EchoRequest",
            fields = listOf(
              Field(
                name = "name",
                type = Type(string = xyz.block.ftl.v1.schema.String())
              )
            ),
            comments = listOf(
              """/**
         * Request to echo a message.
         */"""
            )
          ),
        ),
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
                    key = xyz.block.ftl.v1.schema.Type(string = xyz.block.ftl.v1.schema.String()),
                    value_ = xyz.block.ftl.v1.schema.Type(dataRef = DataRef(name = "MapValue", module = "echo"))
                  )
                )
              )
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
                    element = xyz.block.ftl.v1.schema.Type(
                      dataRef = DataRef(name = "EchoMessage", module = "echo")
                    )
                  )
                )
              )
            )
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
            request = DataRef(
              name = "EchoRequest",
              module = "echo"
            ),
            response = DataRef(
              name = "EchoResponse",
              module = "echo"
            ),
            metadata = listOf(
              Metadata(
                ingress = MetadataIngress(
                  method = "GET",
                  path = "/echo"
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
        )
      )
    )

    assertEquals(expected, module)
  }

  @Test
  fun `fails if invalid schema type is included`() {
    val code = """
        package ftl.echo

        import ftl.time.TimeModuleClient
        import ftl.time.TimeRequest
        import ftl.time.TimeResponse
        import xyz.block.ftl.Context
        import xyz.block.ftl.Ingress
        import xyz.block.ftl.Method
        import xyz.block.ftl.Verb

        class InvalidInput(val field: String) : Exception()

        data class EchoMessage(val message: String, val metadata: Map<String, String>? = null)

        /**
         * Request to echo a message.
         */
        data class EchoRequest(val name: Any)
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
            @Ingress(Method.GET, "/echo")
            fun echo(context: Context, req: EchoRequest): EchoResponse {
                callTime(context)
                return EchoResponse(messages = listOf(EchoMessage(message = "Hello!")))
            }

            fun callTime(context: Context): TimeResponse {
                return context.call(TimeModuleClient::time, TimeRequest())
            }
        }
        """
    assertThrows<IllegalArgumentException>(message = "kotlin.Any type is not supported in FTL schema") {
      ExtractSchemaRule(Config.empty).compileAndLintWithContext(env, code)
    }
  }

  @AfterTest
  fun cleanup() {
    val file = File(OUTPUT_FILENAME)
    file.delete()
  }
}
