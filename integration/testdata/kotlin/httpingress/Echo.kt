package ftl.echo

import ftl.builtin.Empty
import ftl.builtin.HttpRequest
import ftl.builtin.HttpResponse
import kotlin.String
import kotlin.Unit
import xyz.block.ftl.Json
import xyz.block.ftl.Context
import xyz.block.ftl.HttpIngress
import xyz.block.ftl.Method
import xyz.block.ftl.Export

data class GetRequest(
  @Json("userId") val userID: String,
  @Json("postId") val postID: String,
)

data class Nested(
  @Json("good_stuff") val goodStuff: String,
)

data class GetResponse(
  @Json("msg") val message: String,
  @Json("nested") val nested: Nested,
)

data class PostRequest(
  @Json("user_id") val userId: Int,
  val postId: Int,
)

data class PostResponse(
  @Json("success") val success: Boolean,
)

data class PutRequest(
  @Json("userId") val userID: String,
  @Json("postId") val postID: String,
)

data class DeleteRequest(
  @Json("userId") val userID: String,
)

data class ArrayType(
  @Json("item") val item: String,
)

@Export
@HttpIngress(Method.GET, "/users/{userID}/posts/{postID}")
fun `get`(context: Context, req: HttpRequest<GetRequest>): HttpResponse<GetResponse, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("Get" to arrayListOf("Header from FTL")),
    body = GetResponse(
      message = "UserID: ${req.body.userID}, PostID: ${req.body.postID}",
      nested = Nested(goodStuff = "This is good stuff")
    )
  )
}

@Export
@HttpIngress(Method.POST, "/users")
fun post(context: Context, req: HttpRequest<PostRequest>): HttpResponse<PostResponse, String> {
  return HttpResponse(
    status = 201,
    headers = mapOf("Post" to arrayListOf("Header from FTL")),
    body = PostResponse(success = true)
  )
}

@Export
@HttpIngress(Method.PUT, "/users/{userId}")
fun put(context: Context, req: HttpRequest<PutRequest>): HttpResponse<Empty, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("Put" to arrayListOf("Header from FTL")),
    body = Empty()
  )
}

@Export
@HttpIngress(Method.DELETE, "/users/{userId}")
fun delete(context: Context, req: HttpRequest<DeleteRequest>): HttpResponse<Empty, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("Delete" to arrayListOf("Header from FTL")),
    body = Empty()
  )
}

@Export
@HttpIngress(Method.GET, "/html")
fun html(context: Context, req: HttpRequest<Empty>): HttpResponse<String, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("Content-Type" to arrayListOf("text/html; charset=utf-8")),
    body = "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>",
  )
}

@Export
@HttpIngress(Method.POST, "/bytes")
fun bytes(context: Context, req: HttpRequest<ByteArray>): HttpResponse<ByteArray, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("Content-Type" to arrayListOf("application/octet-stream")),
    body = req.body,
  )
}

@Export
@HttpIngress(Method.GET, "/empty")
fun empty(context: Context, req: HttpRequest<Unit>): HttpResponse<Unit, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("Empty" to arrayListOf("Header from FTL")),
    body = Unit
  )
}

@Export
@HttpIngress(Method.GET, "/string")
fun string(context: Context, req: HttpRequest<String>): HttpResponse<String, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("String" to arrayListOf("Header from FTL")),
    body = req.body
  )
}

@Export
@HttpIngress(Method.GET, "/int")
fun int(context: Context, req: HttpRequest<Int>): HttpResponse<Int, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("Int" to arrayListOf("Header from FTL")),
    body = req.body
  )
}

@Export
@HttpIngress(Method.GET, "/float")
fun float(context: Context, req: HttpRequest<Double>): HttpResponse<Double, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("Float" to arrayListOf("Header from FTL")),
    body = req.body
  )
}

@Export
@HttpIngress(Method.GET, "/bool")
fun bool(context: Context, req: HttpRequest<Boolean>): HttpResponse<Boolean, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("Bool" to arrayListOf("Header from FTL")),
    body = req.body
  )
}

@Export
@HttpIngress(Method.GET, "/error")
fun error(context: Context, req: HttpRequest<Unit>): HttpResponse<Boolean, String> {
  return HttpResponse(
    status = 500,
    headers = mapOf("Error" to arrayListOf("Header from FTL")),
    error = "Error from FTL"
  )
}

@Export
@HttpIngress(Method.GET, "/array/string")
fun arrayString(context: Context, req: HttpRequest<List<String>>): HttpResponse<List<String>, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("ArrayString" to arrayListOf("Header from FTL")),
    body = req.body
  )
}

@Export
@HttpIngress(Method.POST, "/array/data")
fun arrayData(context: Context, req: HttpRequest<List<ArrayType>>): HttpResponse<List<ArrayType>, String> {
  return HttpResponse(
    status = 200,
    headers = mapOf("ArrayData" to arrayListOf("Header from FTL")),
    body = req.body
  )
}
