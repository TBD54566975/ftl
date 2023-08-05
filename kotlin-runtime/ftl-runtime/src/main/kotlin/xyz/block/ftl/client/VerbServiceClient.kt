package xyz.block.ftl.client

import com.squareup.wire.GrpcClient
import okio.ByteString.Companion.encodeUtf8
import xyz.block.ftl.Context
import xyz.block.ftl.registry.Registry
import xyz.block.ftl.registry.VerbRef
import xyz.block.ftl.v1.CallRequest
import xyz.block.ftl.v1.VerbServiceClient as VerbServiceClientProto

/**
 * Client for calling verbs. Concrete implementations of this interface may call via gRPC or directly.
 */
interface VerbServiceClient {
  /**
   * Call a verb.
   *
   * @param ref The verb to call.
   * @param req The request encoded as JSON.
   * @return The response encoded as JSON.
   */
  fun call(context: Context, ref: VerbRef, req: String): String
}

class GrpcVerbServiceClient(grpcClient: GrpcClient) : VerbServiceClient {
  val client = grpcClient.create(VerbServiceClientProto::class)

  override fun call(context: Context, ref: VerbRef, req: String): String {
    val request = CallRequest(
      verb = xyz.block.ftl.v1.schema.VerbRef(
        module = ref.module,
        name = ref.name
      ),
      body = req.encodeUtf8(),
    )
    val response = client.Call().executeBlocking(request)
    return when {
      response.error != null -> throw RuntimeException(response.error.message)
      response.body != null -> response.body.utf8()
      else -> error("unreachable")
    }
  }
}

/**
 * A client that calls verbs directly via the associated registry.
 */
class LoopbackVerbServiceClient(private val registry: Registry) : VerbServiceClient {
  override fun call(context: Context, ref: VerbRef, req: String): String {
    return registry.invoke(context, ref, req)
  }
}
