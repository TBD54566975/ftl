package ftl.echo

import kotlinx.datetime.TimeZone
import xyz.block.ftl.Context
import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb
import java.time.Instant

data class EchoRequest(val name: String)
data class EchoResponse(val message: String)

fun time(context: Context, req: EchoRequest): EchoResponse {
  return EchoResponse(message = "Hello, ${req.name}! The time is ${Instant.now()}.")
}

class Echo {
  @Verb @Ingress(Method.GET, "/echo")
  fun echo(context: Context, req: EchoRequest): EchoResponse {
    val tz = TimeZone.currentSystemDefault()
    return EchoResponse(message = "Hello, ${req.name}! The time is ${Instant.now()} in ${tz}.")
  }
}
