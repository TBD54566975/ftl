package schema

// BuiltinsSource is the schema source code for built-in types.
var BuiltinsSource = `
// Built-in types for FTL.
builtin module builtin {
  // HTTP request structure used for HTTP ingress verbs.
  data HttpRequest {
    method String
    path String
    pathParameters {String: String}
    query {String: [String]}
    headers {String: [String]}
    body Bytes
  }

  // HTTP response structure used for HTTP ingress verbs.
  data HttpResponse {
    status Int
    headers {String: [String]}
    body Bytes
  }
}
`

// Builtins returns a [Module] containing built-in types.
func Builtins() *Module {
	module, err := ParseModuleString("builtins.ftl", BuiltinsSource)
	if err != nil {
		panic("failed to parse builtins: " + err.Error())
	}
	return module
}
