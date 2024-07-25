package ftl.time

import ftl.builtin.Empty
import xyz.block.ftl.Context
import xyz.block.ftl.Export
import java.time.OffsetDateTime

data class TimeResponse(val time: OffsetDateTime)

@Export
fun time(context: Context, req: Empty): TimeResponse {
  return TimeResponse(time = OffsetDateTime.now())
}
