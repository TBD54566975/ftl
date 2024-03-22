package ftl.echo

import xyz.block.ftl.Context
import xyz.block.ftl.Method
import xyz.block.ftl.Export

data class EchoRequest(val name: String? = "anonymous")
data class EchoResponse(val message: String)

@Export
fun echo(context: Context, req: EchoRequest): EchoResponse {
  return EchoResponse(message = "Hello, ${req.name}!")
}
