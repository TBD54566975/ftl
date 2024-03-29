package xyz.block.ftl.server

import io.grpc.stub.StreamObserver
import okio.ByteString.Companion.encodeUtf8
import xyz.block.ftl.Context
import xyz.block.ftl.client.VerbServiceClient
import xyz.block.ftl.logging.Logging
import xyz.block.ftl.registry.Registry
import xyz.block.ftl.registry.defaultJvmModuleName
import xyz.block.ftl.registry.toModel
import xyz.block.ftl.v1.*

/**
 * FTL verb server.
 */
class Server(
  val registry: Registry,
  val routingClient: VerbServiceClient,
  val jvmModule: String = defaultJvmModuleName,
) : VerbServiceWireGrpc.VerbServiceImplBase() {

  private val logger = Logging.logger(Server::class)

  override fun Ping(request: PingRequest, response: StreamObserver<PingResponse>) {
    response.onNext(PingResponse())
    response.onCompleted()
  }

  override fun Call(request: CallRequest, response: StreamObserver<CallResponse>) {
    val verbRef = request.verb
    if (verbRef == null) {
      response.onError(IllegalArgumentException("verb is required"))
      return
    }

    try {
      val out = registry.invoke(
        Context(jvmModule, routingClient),
        verbRef.toModel(),
        request.body.utf8()
      )
      response.onNext(CallResponse(body = out.encodeUtf8()))
    } catch (t: Throwable) {
      val stackTrace = t.stackTraceToString()
      logger.error("error calling verb $verbRef: $stackTrace")
      response.onNext(
        CallResponse(
          error = CallResponse.Error(
            message = (t.message ?: stackTrace.lineSequence().firstOrNull() ?: "internal error"),
            stack = stackTrace
          )
        ),
      )
    }
    response.onCompleted()
  }
}
