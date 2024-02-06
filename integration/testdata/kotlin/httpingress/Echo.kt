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
  @Alias("user_id") val userID: Int,
  val postID: Int,
)

data class PostResponse(
  @Alias("success") val success: Boolean,
)

data class PutRequest(
  @Alias("userId") val userID: String,
  @Alias("postId") val postID: String,
)

typealias PutResponse = Unit

data class DeleteRequest(
  @Alias("userId") val userID: String,
)
typealias DeleteResponse = Unit

typealias HtmlRequest = Unit

class Echo {
  @Verb
  @Ingress(
    Method.GET, "/users/{userID}/posts/{postID}", HTTP)
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
  @Ingress(Method.POST, "/users", HTTP)
  fun post(context: Context, req: HttpRequest<PostRequest>): HttpResponse<PostResponse> {
    return HttpResponse<PostResponse>(
      status = 201,
      headers = mapOf("Post" to arrayListOf("Header from FTL")),
      body = PostResponse(success = true)
    )
  }

  @Verb
  @Ingress(Method.PUT, "/users/{userId}", HTTP)
  fun put(context: Context, req: HttpRequest<PutRequest>): HttpResponse<PutResponse> {
    return HttpResponse<PutResponse>(
      status = 200,
      headers = mapOf("Put" to arrayListOf("Header from FTL")),
      body = PutResponse
    )
  }

  @Verb
  @Ingress(Method.DELETE, "/users/{userId}", HTTP)
  fun delete(context: Context, req: HttpRequest<DeleteRequest>): HttpResponse<DeleteResponse> {
    return HttpResponse<DeleteResponse>(
      status = 200,
      headers = mapOf("Delete" to arrayListOf("Header from FTL")),
      body = DeleteResponse
    )
  }

  @Verb
  @Ingress(Method.GET, "/html", HTTP)
  fun html(context: Context, req: HttpRequest<HtmlRequest>): HttpResponse<String> {
    return HttpResponse<String>(
      status = 200,
      headers = mapOf("Content-Type" to arrayListOf("text/html; charset=utf-8")),
      body = "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>",
    )
  }

  @Verb
  @Ingress(Method.POST, "/bytes", HTTP)
  fun bytes(context: Context, req: HttpRequest<ByteArray>): HttpResponse<ByteArray> {
    return HttpResponse<ByteArray>(
      status = 200,
      headers = mapOf("Content-Type" to arrayListOf("application/octet-stream")),
      body = req.body,
    )
  }

  @Verb
  @Ingress(Method.GET, "/empty", HTTP)
  fun empty(context: Context, req: HttpRequest<Unit>): HttpResponse<Unit> {
    return HttpResponse<Unit>(
      status = 200,
      headers = mapOf("Empty" to arrayListOf("Header from FTL")),
      body = Unit
    )
  }

  @Verb
  @Ingress(Method.GET, "/string", HTTP)
  fun string(context: Context, req: HttpRequest<String>): HttpResponse<String> {
    return HttpResponse<String>(
      status = 200,
      headers = mapOf("String" to arrayListOf("Header from FTL")),
      body = req.body
    )
  }
}
