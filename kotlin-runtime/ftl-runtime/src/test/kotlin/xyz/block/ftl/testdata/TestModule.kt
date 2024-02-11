package ftl.test

import ftl.builtin.Empty
import xyz.block.ftl.*
import java.time.OffsetDateTime


data class EchoRequest(val user: String)
data class EchoResponse(val text: String)

@Verb
fun echo(context: Context, req: EchoRequest): EchoResponse {
  val time = context.call(::time, Empty())
  return EchoResponse("Hello ${req.user}, the time is ${time.time}!")
}

data class TimeResponse(val time: OffsetDateTime)

val staticTime = OffsetDateTime.now()

@Verb
fun time(context: Context, req: Empty): TimeResponse {
  return TimeResponse(staticTime)
}

data class VerbRequest(val text: String = "")
data class VerbResponse(val text: String = "")

@Verb
@HttpIngress(Method.GET, "/test")
fun verb(context: Context, req: VerbRequest): VerbResponse {
  return VerbResponse("test")
}


@Verb
@Ignore
fun anotherVerb(context: Context, req: VerbRequest): VerbResponse {
  return VerbResponse("ignored")
}
