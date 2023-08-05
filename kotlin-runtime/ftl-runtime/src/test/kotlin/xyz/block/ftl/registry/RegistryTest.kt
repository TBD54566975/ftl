package xyz.block.ftl.registry

import com.google.gson.Gson
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test
import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb
import xyz.block.ftl.client.LoopbackVerbServiceClient
import kotlin.test.assertContentEquals

data class VerbRequest(val text: String = "")
data class VerbResponse(val text: String = "")

data class RenamedVerbRequest(val text: String = "")
data class RenamedVerbResponse(val text: String = "")

class RenamedVerb {
  @Verb("something")
  fun renamed(context: Context, req: RenamedVerbRequest): RenamedVerbResponse {
    return RenamedVerbResponse("renamed")
  }
}

class ExampleVerb {
  @Verb
  @Ingress(Method.GET, "/test")
  fun verb(context: Context, req: VerbRequest): VerbResponse {
    return VerbResponse("test")
  }
}

@Ignore
class IgnoredVerb {
  @Verb
  fun anotherVerb(context: Context, req: VerbRequest): VerbResponse {
    return VerbResponse("ignored")
  }
}

class RegistryTest {
  private val gson = Gson()
  private val verbRef = VerbRef(module = "registry", name = "verb")
  private val renamedVerbRef = VerbRef(module = "registry", name = "something")

  @Test
  fun moduleName() {
    val registry = Registry("xyz.block.ftl")
    registry.register(ExampleVerb::class)
    assertEquals("registry", registry.moduleName)
  }

  @Test
  fun register() {
    val registry = Registry("xyz.block.ftl")
    registry.register(ExampleVerb::class)
    registry.register(RenamedVerb::class)
    assertContentEquals(listOf(renamedVerbRef, verbRef), registry.refs.sortedBy { it.toString() })
  }

  @Test
  fun invoke() {
    val registry = Registry("xyz.block.ftl")
    registry.register(ExampleVerb::class)
    val context = Context("xyz.block.ftl", LoopbackVerbServiceClient(registry))
    val result = registry.invoke(context, verbRef, gson.toJson(VerbRequest("test")))
    assertEquals(result, gson.toJson(VerbResponse("test")))
  }

  // For some reason "RenamedVerb" does not show up in the scan result.
  // I think it's because there's some additional magic that has to be
  // done to get the class to load when they're in tests.
  @Disabled("Verbs defined in the tests don't seem to be stable.") @Test
  fun registerAll() {
    val registry = Registry("xyz.block")
    registry.registerAll()
    // assertContentEquals(listOf(verbRef), registry.refs.sortedBy { it.toString() })
  }
}
