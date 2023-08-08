package xyz.block.ftl.registry

import org.junit.jupiter.api.Test
import xyz.block.ftl.serializer.makeGson

class VerbHandleTest {
  val gson = makeGson()

  @Test
  fun testInvoke() {
    // val requestJson = gson.toJson(VerbRequest("a"))
    // val handle = VerbHandle(
    //   verbClass = ExampleVerb::class,
    //   verbFunction = ExampleVerb::verb,
    // )
    // val response = handle.invokeVerbInternal(
    //   context = Context(),
    //   argument = requestJson,
    // )
    // assertEquals(gson.toJson(VerbResponse("test")), response)
  }
}
