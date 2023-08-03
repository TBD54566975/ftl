package ftl.echo

import ftl.time.Time
import ftl.time.TimeRequest
import kotlinx.datetime.Instant
import xyz.block.ftl.Context
import xyz.block.ftl.Verb

data class EchoRequest(val name: String)
data class EchoResponse(val message: String)

class Echo {
  @Verb
  fun echo(context: Context, req: EchoRequest): EchoResponse {
    val response = context.call(Time::time, TimeRequest())
    val time = Instant.fromEpochSeconds(response.time.toLong(), 0)
    return EchoResponse(message = "Hello, ${req.name}! The time is ${time}.")
  }
}
