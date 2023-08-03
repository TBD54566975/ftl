package xyz.block.ftl.client

import com.squareup.wire.GrpcClient
import okhttp3.OkHttpClient
import okhttp3.Protocol
import java.time.Duration

internal fun makeGrpcClient(endpoint: String): GrpcClient {
  return GrpcClient.Builder()
    .client(
      OkHttpClient.Builder()
        .readTimeout(Duration.ofSeconds(10))
        .writeTimeout(Duration.ofSeconds(10))
        .callTimeout(Duration.ofSeconds(10))
        .protocols(listOf(Protocol.H2_PRIOR_KNOWLEDGE))
        .build()
    )
    .baseUrl(endpoint)
    .build()
}
