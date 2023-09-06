package xyz.block.ftl.gradle

import com.squareup.wire.GrpcClient
import kotlinx.coroutines.channels.ReceiveChannel
import kotlinx.coroutines.channels.SendChannel
import okhttp3.OkHttpClient
import okhttp3.Protocol
import xyz.block.ftl.v1.ControllerServiceClient
import xyz.block.ftl.v1.GetSchemaRequest
import xyz.block.ftl.v1.PullSchemaRequest
import xyz.block.ftl.v1.PullSchemaResponse
import xyz.block.ftl.v1.schema.Schema
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

  fun getSchema(): Schema? {
    val schemas = mutableListOf<PullSchemaResponse>()
    val client = grpcClient.create(ControllerServiceClient::class)
    return client.GetSchema().executeBlocking(GetSchemaRequest()).schema
  }
}
