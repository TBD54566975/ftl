package ftl.echo

import ftl.time.Time
import ftl.time.TimeRequest
import xyz.block.ftl.Context
import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb

data class EchoRequest(val name: String)

data class EchoResponse(val message: String)

class Echo {
  @Verb
  @Ingress(Method.GET, "/echo")
  fun echo(context: Context, req: EchoRequest): EchoResponse {
    val response = context.call(Time::time, TimeRequest())
    return EchoResponse(message = "Hello, ${req.name}! The time is ${response.time}.")
  }
}
