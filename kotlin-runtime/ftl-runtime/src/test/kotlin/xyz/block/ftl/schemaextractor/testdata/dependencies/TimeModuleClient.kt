package ftl.time

import ftl.builtin.Empty
import xyz.block.ftl.Context
import xyz.block.ftl.HttpIngress
import xyz.block.ftl.Method.GET
import xyz.block.ftl.Export
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
@Export
@HttpIngress(
  GET,
  "/time",
)
fun time(context: Context, req: Empty): TimeResponse =
  throw NotImplementedError("Verb stubs should not be called directly, instead use context.call(TimeModuleClient::time, ...)")

@Export
fun other(context: Context, req: Empty): TimeResponse =
  throw NotImplementedError("Verb stubs should not be called directly, instead use context.call(TimeModuleClient::time, ...)")
