package ftl.test

import ftl.builtin.Empty
import xyz.block.ftl.*
import java.time.OffsetDateTime


data class EchoRequest(val user: String)
data class EchoResponse(val text: String)

@Export
fun echo(context: Context, req: EchoRequest): EchoResponse {
  val time = context.call(::time, Empty())
  return EchoResponse("Hello ${req.user}, the time is ${time.time}!")
}

data class TimeResponse(val time: OffsetDateTime)

val staticTime = OffsetDateTime.now()

@Export
fun time(context: Context, req: Empty): TimeResponse {
  return TimeResponse(staticTime)
}

data class VerbRequest(val text: String = "")
data class VerbResponse(val text: String = "")

@Export
@HttpIngress(Method.GET, "/test")
fun verb(context: Context, req: VerbRequest): VerbResponse {
  return VerbResponse("test")
}


@Export
@Ignore
fun anotherVerb(context: Context, req: VerbRequest): VerbResponse {
  return VerbResponse("ignored")
}
