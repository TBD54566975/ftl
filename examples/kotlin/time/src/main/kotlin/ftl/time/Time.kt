package ftl.time

import xyz.block.ftl.Export
import xyz.block.ftl.Verb
import java.time.OffsetDateTime

data class TimeResponse(val time: OffsetDateTime)

@Verb
@Export
fun time(): TimeResponse {
  return TimeResponse(time = OffsetDateTime.now())
}
