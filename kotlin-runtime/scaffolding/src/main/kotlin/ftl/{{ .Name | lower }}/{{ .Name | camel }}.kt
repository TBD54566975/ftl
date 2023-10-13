package ftl.{{ .Name | lower }}

import xyz.block.ftl.Context
import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb

data class {{ .Name | camel }}Request(val name: String)
data class {{ .Name | camel }}Response(val message: String)

class {{ .Name | camel }} {
  @Verb
  fun {{ .Name | lower }}(context: Context, req: {{ .Name | camel }}Request): {{ .Name | camel }}Response {
    return {{ .Name | camel }}Response(message = "Hello, ${req.name}!")
  }
}

