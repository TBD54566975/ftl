package ftl.echo

import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb
import java.time.Instant

data class EchoRequest(val name: String)
data class EchoResponse(val message: String)

class Echo {
  @Verb @Ingress(Method.GET, "/echo")
  fun echo(req: EchoRequest): EchoResponse {
    return EchoResponse(message = "Hello, ${req.name}! The time is ${Instant.now()}.")
  }
}