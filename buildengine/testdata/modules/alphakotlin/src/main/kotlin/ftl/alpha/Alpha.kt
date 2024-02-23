package ftl.alpha

import ftl.builtin.Empty
import ftl.other.Other
import xyz.block.ftl.Context
import xyz.block.ftl.Verb

data class EchoRequest(val name: String?)
data class EchoResponse(val message: String)

@Verb
fun echo(context: Context, req: EchoRequest): EchoResponse {
  val other = Other()
  return EchoResponse(message = "Hello, ${req.name ?: "anonymous"}!.")
}
