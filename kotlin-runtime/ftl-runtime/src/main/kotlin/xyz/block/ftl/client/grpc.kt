package xyz.block.ftl.client

import io.grpc.*
import io.grpc.netty.NettyChannelBuilder
import io.grpc.netty.NettyServerBuilder
import xyz.block.ftl.logging.Logging
import xyz.block.ftl.server.ServerInterceptor
import java.net.InetSocketAddress
import java.net.URL
import java.util.concurrent.TimeUnit.SECONDS

internal fun makeGrpcClient(endpoint: String): ManagedChannel {
  val url = URL(endpoint)
  val port = if (url.port == -1) when (url.protocol) {
    "http" -> 80
    "https" -> 443
    else -> throw IllegalArgumentException("Unsupported protocol: ${url.protocol}")
  } else url.port
  var builder = NettyChannelBuilder
    .forAddress(InetSocketAddress(url.host, port))
    .keepAliveTime(5, SECONDS)
    .intercept(VerbServiceClientInterceptor())
  if (url.protocol == "http")
    builder = builder.usePlaintext()
  return builder.build()
}

private class VerbServiceClientInterceptor : ClientInterceptor {
  override fun <ReqT : Any?, RespT : Any?> interceptCall(
    method: MethodDescriptor<ReqT, RespT>?,
    callOptions: CallOptions?,
    next: Channel?
  ): ClientCall<ReqT, RespT> {
    val call = next?.newCall(method, callOptions)
    return object : ForwardingClientCall.SimpleForwardingClientCall<ReqT, RespT>(call) {
      override fun start(responseListener: Listener<RespT>?, headers: Metadata?) {
        ServerInterceptor.callers.get().forEach { caller ->
          headers?.put(ServerInterceptor.callersMetadata, caller)
        }
        headers?.put(ServerInterceptor.requestIdMetadata, ServerInterceptor.requestId.get())
        super.start(responseListener, headers)
      }
    }
  }

}
