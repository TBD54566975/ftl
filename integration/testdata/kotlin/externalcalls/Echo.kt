package ftl.echo

import ftl.echo2.Echo2ModuleClient
import xyz.block.ftl.Context
import xyz.block.ftl.Verb

data class EchoRequest(val name: String)
data class EchoResponse(val message: String)

class Echo {
  @Verb
  fun echo(context: Context, req: EchoRequest): EchoResponse {
    return EchoResponse(message = "Hello, ${req.name}!")
  }

  @Verb
  fun call(context: Context, req: EchoRequest): EchoResponse {
    val res = context.call(Echo2ModuleClient::echo, ftl.echo2.EchoRequest(name = req.name))
    return EchoResponse(message = res.message)
  }
}
