package ftl.echo

import ftl.builtin.Empty
import ftl.time.time
import xyz.block.ftl.Context
import xyz.block.ftl.Verb

class InvalidInput(val field: String) : Exception()

data class EchoRequest(val name: String?)
data class EchoResponse(val message: String)

@Throws(InvalidInput::class)
@Verb
fun echo(context: Context, req: EchoRequest): EchoResponse {
  val response = context.call(::time, Empty())
  return EchoResponse(message = "Hello, ${req.name ?: "anonymous"}! The time is ${response.time}.")
}
