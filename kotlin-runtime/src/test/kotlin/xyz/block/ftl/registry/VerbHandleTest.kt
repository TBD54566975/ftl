package xyz.block.ftl.registry

import com.google.gson.Gson
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import xyz.block.ftl.Context
import xyz.block.ftl.v1.schema.VerbRef

class VerbHandleTest {
  val gson = Gson()

  @Test
  fun testInvoke() {
    val requestJson = gson.toJson(VerbRequest("a"))
    val handle = VerbHandle(
      verbRef = VerbRef(name = "verb"),
      verbClass = ExampleVerb::class,
      verbFunction = ExampleVerb::verb,
    )
    val response = handle.invokeVerbInternal(
      context = Context(),
      argument = requestJson,
    )
    assertEquals(gson.toJson(VerbResponse("test")), response)
  }
}