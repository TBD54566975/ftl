package xyz.block.ftl

import ftl.builtin.Empty
import ftl.test.*
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.MethodSource
import xyz.block.ftl.client.LoopbackVerbServiceClient
import xyz.block.ftl.registry.Registry

data class TestCase(val expected: Any, val invoke: (ctx: Context) -> Any)

class ContextTest {
  companion object {
    @JvmStatic
    fun endToEnd(): List<TestCase> {
      return listOf(
        TestCase(
          invoke = { ctx -> ctx.call(::echo, EchoRequest("Alice")) },
          expected = EchoResponse("Hello Alice, the time is $staticTime!"),
        ),
        TestCase(
          invoke = { ctx -> ctx.call(::time, Empty()) },
          expected = TimeResponse(staticTime),
        ),
        TestCase(
          invoke = { ctx -> ctx.callSink(::time, EchoRequest("Alice")) },
          expected = Unit,
        ),
        TestCase(
          invoke = { ctx -> ctx.callSource(::time) },
          expected = TimeResponse(staticTime),
        ),
        TestCase(
          invoke = { ctx -> ctx.callEmpty(::time) },
          expected = Unit,
        ),
      )
    }
  }

  @ParameterizedTest
  @MethodSource
  fun endToEnd(testCase: TestCase) {
    val registry = Registry("ftl.test")
    registry.registerAll()
    val routingClient = LoopbackVerbServiceClient(registry)
    val context = Context("ftl.test", routingClient)
    val result = testCase.invoke(context)
    assertEquals(testCase.expected, result)
  }
}
