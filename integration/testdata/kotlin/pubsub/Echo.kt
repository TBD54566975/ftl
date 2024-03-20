package ftl.echo

import ftl.builtin.Empty
import xyz.block.ftl.Context
import xyz.block.ftl.Export

data class EchoResponse(val name: String)

@Export
fun echo(context: Context, req: Empty): EchoResponse {
  context.callSink(::sink, Empty())
  val resp = context.callSource(::source)
  return EchoResponse(name = resp.name)
}

data class SourceResponse(val name: String)

@Export
fun source(context: Context): EchoResponse {
  return EchoResponse(name = "source")
}

@Export
fun sink(context: Context, req: Empty) {
}
