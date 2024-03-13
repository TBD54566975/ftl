package xyz.block.ftl.registry

import ftl.test.VerbRequest
import ftl.test.VerbResponse
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import xyz.block.ftl.Context
import xyz.block.ftl.client.LoopbackVerbServiceClient
import xyz.block.ftl.serializer.makeGson
import kotlin.test.assertContentEquals

class RegistryTest {
  private val gson = makeGson()
  private val verbRef = Ref(module = "test", name = "verb")

  @Test
  fun moduleName() {
    val registry = Registry("ftl.test")
    registry.registerAll()
    assertEquals("test", registry.moduleName)
  }

  @Test
  fun registerAll() {
    val registry = Registry("ftl.test")
    registry.registerAll()
    assertContentEquals(listOf(
      Ref(module = "test", name = "echo"),
      Ref(module = "test", name = "time"),
      Ref(module = "test", name = "verb"),
    ), registry.refs.sortedBy { it.toString() })
  }

  @Test
  fun invoke() {
    val registry = Registry("ftl.test")
    registry.registerAll()
    val context = Context("ftl.test", LoopbackVerbServiceClient(registry))
    val result = registry.invoke(context, verbRef, gson.toJson(VerbRequest("test")))
    assertEquals(result, gson.toJson(VerbResponse("test")))
  }
}
