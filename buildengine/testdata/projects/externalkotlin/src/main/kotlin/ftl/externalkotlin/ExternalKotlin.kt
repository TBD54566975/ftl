package ftl.externalkotlin

import com.google.type.DayOfWeek
import xyz.block.ftl.Context
import xyz.block.ftl.Export
import xyz.block.ftl.Visibility
import xyz.block.ftl.v1.schema.Optional

class InvalidInput(val field: String) : Exception()

data class ExternalRequest(val name: String?, val dayOfWeek: DayOfWeek)
data class ExternalResponse(val message: String)

@Throws(InvalidInput::class)
@Export(Visibility.INTERNAL)
fun external(context: Context, req: ExternalRequest): ExternalResponse {
  return ExternalResponse(message = "Hello, ${req.name ?: "anonymous"}!")
}
