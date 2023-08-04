package xyz.block.ftl

import org.junit.jupiter.api.Test
import xyz.block.ftl.client.LoopbackVerbServiceClient
import xyz.block.ftl.registry.Registry
import kotlin.test.assertEquals

data class Request(val who: String)
data class Response(val message: String)

class Module {
  @Verb fun verb(context: Context, req: Request): Response {
    return Response("Hello, ${req.who}!")
  }
}

class ContextTest {
  @Test
  fun call() {
    val registry = Registry(jvmModuleName = "xyz.block")
    registry.register(Module::class)
    val context = Context("xyz.block", LoopbackVerbServiceClient(registry))
    val response = context.call(Module::verb, Request("world"))
    assertEquals(Response("Hello, world!"), response)
  }
}
