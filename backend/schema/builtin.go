package schema

import "golang.design/x/reflect"

// BuiltinsSource is the schema source code for built-in types.
const BuiltinsSource = `
// Built-in types for FTL.
builtin module builtin {
  // HTTP request structure used for HTTP ingress verbs.
  data HttpRequest<Body> {
    method String
    path String
    pathParameters {String: String}
    query {String: [String]}
    headers {String: [String]}
    body Body
  }

  // HTTP response structure used for HTTP ingress verbs.
  data HttpResponse<Body> {
    status Int
    headers {String: [String]}
    body Body
  }

  data Empty {}
}
`

var builtinsModuleParsed = func() *Module {
	module, err := moduleParser.ParseString("", BuiltinsSource)
	if err != nil {
		panic(err)
	}
	return module
}()

// Builtins returns a [Module] containing built-in types.
func Builtins() *Module {
	return reflect.DeepCopy(builtinsModuleParsed)
}
