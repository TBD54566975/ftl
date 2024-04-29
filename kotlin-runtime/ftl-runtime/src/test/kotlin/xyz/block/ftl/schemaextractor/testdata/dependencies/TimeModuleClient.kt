package ftl.time

import ftl.builtin.Empty
import xyz.block.ftl.Context
import xyz.block.ftl.HttpIngress
import xyz.block.ftl.Method.GET
import xyz.block.ftl.Verb
import java.time.OffsetDateTime

data class TimeResponse(
  val time: OffsetDateTime,
)

enum class Color {
  RED,
  GREEN,
  BLUE,
}

/**
 * Time returns the current time.
 */
@HttpIngress(
  GET,
  "/time",
)
fun time(context: Context, req: Empty): TimeResponse =
  throw NotImplementedError("Verb stubs should not be called directly, instead use context.call(TimeModuleClient::time, ...)")

@Verb
fun other(context: Context, req: Empty): TimeResponse =
  throw NotImplementedError("Verb stubs should not be called directly, instead use context.call(TimeModuleClient::time, ...)")
