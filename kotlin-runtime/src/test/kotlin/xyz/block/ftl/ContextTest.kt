package xyz.block.ftl

import org.junit.jupiter.api.Test

data class Request(val name: String)
data class Response(val message: String)

@Ignore
class TestVerb {
  @Verb
  fun test(context: Context, req: Request): Response {
    return Response(message = "Hello, ${req.name}!")
  }
}

fun freeFunction(context: Context, req: Request): Response {
  return Response(message = "Hello, ${req.name}!")
}

class ContextTest {
  @Test
  fun callMethod() {
    val context = Context()
    // This is just to test that everything works.
    val response = context.call(TestVerb::test, Request(name = "World"))
    assert(response.message == "Hello, World!")
  }

  @Test
  fun callFreeFunction() {
    val context = Context()
    // This is just to test that everything works.
    val response = context.call(::freeFunction, Request(name = "World"))
    assert(response.message == "Hello, World!")
  }
}
