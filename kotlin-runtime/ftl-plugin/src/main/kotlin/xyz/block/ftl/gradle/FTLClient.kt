package xyz.block.ftl.gradle

import com.squareup.wire.GrpcClient
import kotlinx.coroutines.channels.ReceiveChannel
import kotlinx.coroutines.channels.SendChannel
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import okhttp3.OkHttpClient
import okhttp3.Protocol
import xyz.block.ftl.v1.ControllerServiceClient
import xyz.block.ftl.v1.PullSchemaRequest
import xyz.block.ftl.v1.PullSchemaResponse
import java.net.ConnectException
import java.time.Duration

class FTLClient(ftlEndpoint: String) {
  private var grpcClient: GrpcClient
  private lateinit var sendSchemaChannel: SendChannel<PullSchemaRequest>
  private lateinit var receiveSchemaChannel: ReceiveChannel<PullSchemaResponse>

  init {
    grpcClient = GrpcClient.Builder()
      .client(
        OkHttpClient.Builder()
          .readTimeout(Duration.ofSeconds(10))
          .writeTimeout(Duration.ofSeconds(10))
          .callTimeout(Duration.ofSeconds(10))
          .protocols(listOf(Protocol.H2_PRIOR_KNOWLEDGE))
          .addInterceptor { chain ->
            try {
              chain.proceed(chain.request())
            } catch (e: ConnectException) {
              throw ConnectException(
                "Unable to connect to FTL Controller at: $ftlEndpoint. Is it running?"
              )
            }
          }
          .build()
      )
      .baseUrl(ftlEndpoint)
      .build()
  }

  fun pullSchemas(): List<PullSchemaResponse> {
    val schemas = mutableListOf<PullSchemaResponse>()
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
          for (schema in receiveSchemaChannel) {
            schemas.add(schema)
            if (!schema.more) {
              receiveSchemaChannel.cancel()
              return@launch
            }
          }
        } catch (e: Exception) {
          receiveSchemaChannel.cancel()
          throw e
        }
      }
    }
    return schemas
  }
}
