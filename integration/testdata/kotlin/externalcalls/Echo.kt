package ftl.echo

import ftl.echo2.echo as echo2
import xyz.block.ftl.Context
import xyz.block.ftl.Verb

data class EchoRequest(val name: String)
data class EchoResponse(val message: String)

@Verb
fun echo(context: Context, req: EchoRequest): EchoResponse {
  return EchoResponse(message = "Hello, ${req.name}!")
}

@Verb
fun call(context: Context, req: EchoRequest): EchoResponse {
  val res = context.call(::echo2, ftl.echo2.EchoRequest(name = req.name))
  return EchoResponse(message = res.message)
}
