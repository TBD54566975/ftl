package ftl.{{ .Name | camel | lower }}

import xyz.block.ftl.Context
import xyz.block.ftl.Method
import xyz.block.ftl.Export
import xyz.block.ftl.Visibility

data class EchoRequest(val name: String? = "anonymous")
data class EchoResponse(val message: String)

@Export(Visibility.INTERNAL)
fun echo(context: Context, req: EchoRequest): EchoResponse {
  return EchoResponse(message = "Hello, ${req.name}!")
}
