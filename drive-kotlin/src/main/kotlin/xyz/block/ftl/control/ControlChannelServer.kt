package xyz.block.ftl.control

import com.google.gson.Gson
import com.google.protobuf.ByteString
import io.grpc.Status
import io.grpc.netty.NettyServerBuilder
import xyz.block.ftl.Context
import xyz.block.ftl.drive.verb.VerbDeck
import xyz.block.ftl.v1.*
import java.net.SocketAddress


class ControlChannelServer(val deck: VerbDeck) : DriveServiceGrpcKt.DriveServiceCoroutineImplBase() {
  private val gson = Gson()

  override suspend fun ping(request: PingRequest): PingResponse {
    return PingResponse.getDefaultInstance()
  }

  override suspend fun call(request: CallRequest): CallResponse {
    val cassette = deck.lookupFullyQualifiedName(request.verb) ?: throw Status.NOT_FOUND.asException()
    val req = gson.fromJson<Any>(request.body.toStringUtf8(), cassette.argumentType.java)
    var resp: Any
    try {
      resp = cassette.dispatch(Context(http = null), req) // TODO: do something with Context
    } catch (e: Exception) {
      return CallResponse.newBuilder()
        .setError(
          CallResponse.Error.newBuilder()
            .setMessage(e.message ?: "no error message")
            .build()
        )
        .build()
    }
    return CallResponse.newBuilder()
      .setBody(ByteString.copyFromUtf8(gson.toJson(resp)))
      .build()
  }

  override suspend fun list(request: ListRequest): ListResponse {
    return ListResponse.newBuilder()
      .addAllVerbs(deck.list().map { deck.fullyQualifiedName(it) })
      .build()
  }

  override suspend fun fileChange(request: FileChangeRequest): FileChangeResponse {
    return FileChangeResponse.getDefaultInstance()
  }
}

/**
 * Start DriveService on the given socket.
 */
fun startControlChannelServer(socket: SocketAddress, controlChannel: ControlChannelServer) {
  val server = NettyServerBuilder
    .forAddress(socket)
    .addService(controlChannel)
    .build()
  // TODO: Terminate the process if this fails to startup.
  server.start()
}