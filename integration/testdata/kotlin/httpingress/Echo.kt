package ftl.echo

import ftl.builtin.Empty
import ftl.builtin.HttpRequest
import ftl.builtin.HttpResponse
import kotlin.String
import kotlin.Unit
import xyz.block.ftl.Alias
import xyz.block.ftl.Context
import xyz.block.ftl.HttpIngress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb

data class GetRequest(
  @Alias("userId") val userID: String,
  @Alias("postId") val postID: String,
)

data class Nested(
  @Alias("good_stuff") val goodStuff: String,
)

data class GetResponse(
  @Alias("msg") val message: String,
  @Alias("nested") val nested: Nested,
)

data class PostRequest(
  @Alias("user_id") val userId: Int,
  val postId: Int,
)

data class PostResponse(
  @Alias("success") val success: Boolean,
)

data class PutRequest(
  @Alias("userId") val userID: String,
  @Alias("postId") val postID: String,
)

data class DeleteRequest(
  @Alias("userId") val userID: String,
)

class Echo {
  @Verb
  @HttpIngress(
    Method.GET, "/users/{userID}/posts/{postID}")
  fun `get`(context: Context, req: HttpRequest<GetRequest>): HttpResponse<GetResponse> {
    return HttpResponse<GetResponse>(
      status = 200,
      headers = mapOf("Get" to arrayListOf("Header from FTL")),
      body = GetResponse(
        message = "UserID: ${req.body.userID}, PostID: ${req.body.postID}",
        nested = Nested(goodStuff = "This is good stuff")
      )
    )
  }

  @Verb
  @HttpIngress(Method.POST, "/users")
  fun post(context: Context, req: HttpRequest<PostRequest>): HttpResponse<PostResponse> {
    return HttpResponse<PostResponse>(
      status = 201,
      headers = mapOf("Post" to arrayListOf("Header from FTL")),
      body = PostResponse(success = true)
    )
  }

  @Verb
  @HttpIngress(Method.PUT, "/users/{userId}")
  fun put(context: Context, req: HttpRequest<PutRequest>): HttpResponse<Empty> {
    return HttpResponse<Empty>(
      status = 200,
      headers = mapOf("Put" to arrayListOf("Header from FTL")),
      body = Empty()
    )
  }

  @Verb
  @HttpIngress(Method.DELETE, "/users/{userId}")
  fun delete(context: Context, req: HttpRequest<DeleteRequest>): HttpResponse<Empty> {
    return HttpResponse<Empty>(
      status = 200,
      headers = mapOf("Delete" to arrayListOf("Header from FTL")),
      body = Empty()
    )
  }

  @Verb
  @HttpIngress(Method.GET, "/html")
  fun html(context: Context, req: HttpRequest<Empty>): HttpResponse<String> {
    return HttpResponse<String>(
      status = 200,
      headers = mapOf("Content-Type" to arrayListOf("text/html; charset=utf-8")),
      body = "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>",
    )
  }

  @Verb
  @HttpIngress(Method.POST, "/bytes")
  fun bytes(context: Context, req: HttpRequest<ByteArray>): HttpResponse<ByteArray> {
    return HttpResponse<ByteArray>(
      status = 200,
      headers = mapOf("Content-Type" to arrayListOf("application/octet-stream")),
      body = req.body,
    )
  }

  @Verb
  @HttpIngress(Method.GET, "/empty")
  fun empty(context: Context, req: HttpRequest<Unit>): HttpResponse<Unit> {
    return HttpResponse<Unit>(
      status = 200,
      headers = mapOf("Empty" to arrayListOf("Header from FTL")),
      body = Unit
    )
  }

  @Verb
  @HttpIngress(Method.GET, "/string")
  fun string(context: Context, req: HttpRequest<String>): HttpResponse<String> {
    return HttpResponse<String>(
      status = 200,
      headers = mapOf("String" to arrayListOf("Header from FTL")),
      body = req.body
    )
  }
}
