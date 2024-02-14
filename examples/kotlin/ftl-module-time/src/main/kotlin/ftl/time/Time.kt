package ftl.time

import ftl.builtin.Empty
import xyz.block.ftl.Context
import xyz.block.ftl.Verb
import java.time.OffsetDateTime

// Make time an OffsetDateTime once schema extraction is fixed
data class TimeResponse(val time: String)

@Verb
fun time(context: Context, req: Empty): TimeResponse {
  return TimeResponse(time = OffsetDateTime.now().toString())
}
