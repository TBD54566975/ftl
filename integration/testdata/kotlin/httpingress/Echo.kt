package ftl.echo

import ftl.builtin.HttpRequest
import ftl.builtin.HttpResponse
import kotlin.String
import kotlin.Unit
import xyz.block.ftl.Alias
import xyz.block.ftl.Context
import xyz.block.ftl.Ingress
import xyz.block.ftl.Ingress.Type.HTTP
import xyz.block.ftl.Method
import xyz.block.ftl.Verb

data class GetRequest(
  val userID: String,
  val postID: String,
)

data class GetResponse(
  val message: String,
)

data class PostRequest(
  @Alias("id") val userID: String,
  val postID: String,
)

data class PostResponse(
  val message: String,
)

data class PutRequest(
  val userID: String,
  val postID: String,
)

data class PutResponse(
  val message: String,
)

data class DeleteRequest(
  val userID: String,
)

data class DeleteResponse(
  val message: String,
)

class Echo {
  @Verb
  @Ingress(
    Method.GET,
    "/echo/users/{userID}/posts/{postID}",
    HTTP
  )
  fun `get`(context: Context, req: HttpRequest<GetRequest>): HttpResponse<GetResponse> {
    return HttpResponse<GetResponse>(
      status = 200,
      headers = mapOf("Get" to arrayListOf("Header from FTL")),
      body = GetResponse(message = "UserID: ${req.body.userID}, PostID ${req.body.postID}")
    )
  }

  @Verb
  @Ingress(
    Method.POST,
    "/echo/users",
    HTTP
  )
  fun post(context: Context, req: HttpRequest<PostRequest>): HttpResponse<PostResponse> {
    return HttpResponse<PostResponse>(
      status = 201,
      headers = mapOf("Post" to arrayListOf("Header from FTL")),
      body = PostResponse(message = "UserID: ${req.body.userID}, PostID ${req.body.postID}")
    )
  }

  @Verb
  @Ingress(
    Method.PUT,
    "/echo/users/{userID}",
    HTTP
  )
  fun put(context: Context, req: HttpRequest<PutRequest>): HttpResponse<PutResponse> {
    return HttpResponse<PutResponse>(
      status = 200,
      headers = mapOf("Put" to arrayListOf("Header from FTL")),
      body = PutResponse(message = "UserID: ${req.body.userID}, PostID ${req.body.postID}")
    )
  }

  @Verb
  @Ingress(Method.DELETE, "/echo/users/{userID}", HTTP)
  fun delete(context: Context, req: HttpRequest<DeleteRequest>): HttpResponse<DeleteResponse> {
    return HttpResponse<DeleteResponse>(
      status = 200,
      headers = mapOf("Delete" to arrayListOf("Header from FTL")),
      body = DeleteResponse(message = "UserID: ${req.body.userID}")
    )
  }
}
