package ftl.time

import ftl.builtin.Empty
import xyz.block.ftl.Context
import xyz.block.ftl.Verb
import java.time.OffsetDateTime

data class TimeResponse(val time: OffsetDateTime)

@Verb
fun time(context: Context, req: Empty): TimeResponse {
  return TimeResponse(time = OffsetDateTime.now())
}
