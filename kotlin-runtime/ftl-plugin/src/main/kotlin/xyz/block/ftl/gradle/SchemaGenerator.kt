package xyz.block.ftl.gradle

import com.squareup.kotlinpoet.*
import com.squareup.wire.GrpcClient
import kotlinx.coroutines.*
import kotlinx.coroutines.channels.ReceiveChannel
import kotlinx.coroutines.channels.SendChannel
import okhttp3.OkHttpClient
import okhttp3.Protocol
import xyz.block.ftl.v1.ControllerServiceClient
import xyz.block.ftl.v1.PullSchemaRequest
import xyz.block.ftl.v1.PullSchemaResponse
import java.time.Duration

class SchemaGenerator(ftlEndpoint: String) {
  private var grpcClient: GrpcClient
  private lateinit var sendSchemaChannel: SendChannel<PullSchemaRequest>
  private lateinit var receiveSchemaChannel: ReceiveChannel<PullSchemaResponse>

  init {
    grpcClient = GrpcClient.Builder()
      .client(
        OkHttpClient.Builder()
          .readTimeout(Duration.ofSeconds(5))
          .writeTimeout(Duration.ofSeconds(5))
          .callTimeout(Duration.ofSeconds(5))
          .protocols(listOf(Protocol.H2_PRIOR_KNOWLEDGE))
          .build()
      )
      .baseUrl(ftlEndpoint)
      .build()
  }

  fun generate() {
    runBlocking {
      launch {
        grpcClient.create(ControllerServiceClient::class)
          .PullSchema()
          .executeIn(this)
          .let { (sendChannel, receiveChannel) ->
            sendSchemaChannel = sendChannel
            receiveSchemaChannel = receiveChannel
          }

        require(sendSchemaChannel.trySend(PullSchemaRequest()).isSuccess)
        try {
          println("Trying...")

          for (schema in receiveSchemaChannel) {
            println(schema.schema?.name)
            generateModule()
            if (!schema.more) {
              receiveSchemaChannel.cancel()
              return@launch
            }
          }
          println("Tried")
        } catch (e: Exception) {
          println(e)
        }
      }
    }
  }

  private fun generateModule(schema: PullSchemaResponse) {
    val greeterClass = ClassName("", schema.module_name)
    val file = FileSpec.builder("", "HelloWorld")
      .addType(
        TypeSpec.classBuilder("Greeter")
          .primaryConstructor(
            FunSpec.constructorBuilder()
              .addParameter("name", String::class)
              .build()
          )
          .addProperty(
            PropertySpec.builder("name", String::class)
              .initializer("name")
              .build()
          )
          .addFunction(
            FunSpec.builder("greet")
              .addStatement("println(%P)", "Hello, \$name")
              .build()
          )
          .build()
      )
      .addFunction(
        FunSpec.builder("main")
          .addParameter("args", String::class, KModifier.VARARG)
          .addStatement("%T(args[0]).greet()", greeterClass)
          .build()
      )
      .build()
    file.writeTo(System.out)
  }
}
