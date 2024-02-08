package xyz.block.ftl

import ftl.builtin.Empty
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.MethodSource
import xyz.block.ftl.client.LoopbackVerbServiceClient
import xyz.block.ftl.registry.Registry
import java.time.OffsetDateTime

data class EchoRequest(val user: String)
data class EchoResponse(val text: String)

class Echo {
  @Verb
  fun echo(context: Context, req: EchoRequest): EchoResponse {
    val time = context.call(Time::time, Empty())
    return EchoResponse("Hello ${req.user}, the time is ${time.time}!")
  }
}

data class TimeResponse(val time: OffsetDateTime)

val staticTime = OffsetDateTime.now()

class Time {
  @Verb
  fun time(context: Context, req: Empty): TimeResponse {
    return TimeResponse(staticTime)
  }
}

data class TestCase(val expected: Any, val invoke: (ctx: Context) -> Any)

class ContextTest {
  companion object {
    @JvmStatic
    fun endToEnd(): List<TestCase> {
      return listOf(
        TestCase(
          invoke = { ctx -> ctx.call(Echo::echo, EchoRequest("Alice")) },
          expected = EchoResponse("Hello Alice, the time is $staticTime!"),
        ),
        TestCase(
          invoke = { ctx -> ctx.call(Time::time, Empty()) },
          expected = TimeResponse(staticTime),
        ),
      )
    }
  }

  @ParameterizedTest
  @MethodSource
  fun endToEnd(testCase: TestCase) {
    val registry = Registry("xyz.block")
    registry.register(Echo::class)
    registry.register(Time::class)
    val routingClient = LoopbackVerbServiceClient(registry)
    val context = Context("xyz.block", routingClient)
    val result = testCase.invoke(context)
    assertEquals(result, testCase.expected)
  }
}
