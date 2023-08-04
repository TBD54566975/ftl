package xyz.block.ftl.server

import io.grpc.stub.StreamObserver
import okio.ByteString.Companion.encodeUtf8
import xyz.block.ftl.Context
import xyz.block.ftl.client.VerbServiceClient
import xyz.block.ftl.registry.Registry
import xyz.block.ftl.registry.defaultJvmModuleName
import xyz.block.ftl.registry.toModel
import xyz.block.ftl.v1.CallRequest
import xyz.block.ftl.v1.CallResponse
import xyz.block.ftl.v1.PingRequest
import xyz.block.ftl.v1.PingResponse
import xyz.block.ftl.v1.VerbServiceWireGrpc

/**
 * FTL verb server.
 */
class Server(
  val registry: Registry,
  val routingClient: VerbServiceClient,
  val jvmModule: String = defaultJvmModuleName,
) : VerbServiceWireGrpc.VerbServiceImplBase() {

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
    val out = registry.invoke(
      Context(jvmModule, routingClient),
      verbRef.toModel(),
      request.body.utf8()
    )
    response.onNext(CallResponse(body = out.encodeUtf8()))
    response.onCompleted()
  }
}
