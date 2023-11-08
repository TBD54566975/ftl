package xyz.block.ftl.server

import io.grpc.*
import io.grpc.ServerInterceptor

const val ftlVerbHeader = "FTL-Verb"
const val ftlRequestIdHeader = "FTL-Request-ID"

internal class ServerInterceptor : ServerInterceptor {

  companion object {
    internal var callersMetadata = Metadata.Key.of(ftlVerbHeader, Metadata.ASCII_STRING_MARSHALLER)
    internal var requestIdMetadata = Metadata.Key.of(ftlRequestIdHeader, Metadata.ASCII_STRING_MARSHALLER)

    internal var callers = Context.key<List<String>>(ftlVerbHeader)
    internal var requestId = Context.key<String>(ftlRequestIdHeader)
  }

  override fun <ReqT : Any?, RespT : Any?> interceptCall(
    call: ServerCall<ReqT, RespT>?,
    headers: Metadata?,
    next: ServerCallHandler<ReqT, RespT>?
  ): ServerCall.Listener<ReqT> {
    call?.setCompression("gzip")

    var context = Context.current()

    headers?.getAll(callersMetadata)?.apply {
      context = context.withValue(callers, this.toList())
    }
    headers?.get(requestIdMetadata)?.apply {
      context = context.withValue(requestId, this)
    }

    return Contexts.interceptCall(context, call, headers, next)
  }
}
