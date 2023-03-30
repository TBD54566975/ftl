package xyz.block.ftl.control

import com.google.gson.Gson
import com.google.protobuf.ByteString
import io.grpc.Status
import xyz.block.ftl.Context
import xyz.block.ftl.drive.verb.VerbDeck
import xyz.block.ftl.protoext.fullyQualified
import xyz.block.ftl.v1.CallRequest
import xyz.block.ftl.v1.CallResponse
import xyz.block.ftl.v1.DevelServiceGrpcKt
import xyz.block.ftl.v1.ListRequest
import xyz.block.ftl.v1.ListResponse
import xyz.block.ftl.v1.PingRequest
import xyz.block.ftl.v1.PingResponse
import xyz.block.ftl.v1.VerbServiceGrpcKt
import xyz.block.ftl.v1.schema.VerbRef

class VerbServer(
  private val deck: VerbDeck
) : VerbServiceGrpcKt.VerbServiceCoroutineImplBase() {
  private val gson = Gson()

  override suspend fun ping(request: PingRequest): PingResponse = PingResponse.getDefaultInstance()

  override suspend fun call(request: CallRequest): CallResponse {
    val verb = deck.lookupFullyQualifiedName(request.verb.fullyQualified)
      ?: throw Status.NOT_FOUND.asException()
    val argument = gson.fromJson<Any>(request.body.toStringUtf8(), verb.argumentType.java)
    val reply: Any
    try {
      reply = deck.dispatch(Context.fromAgent(verb.verbId), verb.verbId, argument)
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
      .setBody(ByteString.copyFromUtf8(gson.toJson(reply)))
      .build()
  }

  override suspend fun list(request: ListRequest): ListResponse {
    return ListResponse.newBuilder()
      .addAllVerbs(deck.list().map {
        VerbRef.newBuilder()
          .setModule(deck.module)
          .setName(it.qualifiedName)
          .build()
      })
      .build()
  }
}

class DevelServer() : DevelServiceGrpcKt.DevelServiceCoroutineImplBase() {
  override suspend fun ping(request: PingRequest): PingResponse = PingResponse.getDefaultInstance()
}
