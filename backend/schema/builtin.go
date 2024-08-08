package schema

import "github.com/TBD54566975/ftl/internal/reflect"

// BuiltinsSource is the schema source code for built-in types.
const BuiltinsSource = `
// Built-in types for FTL.
builtin module builtin {
  // Ref is used to reference types, verbs and resources.
  export data Ref {
    module String
    name String
  }
  
  // HTTP request structure used for HTTP ingress verbs.
  export data HttpRequest<Body> {
    method String
    path String
    pathParameters {String: String}
    query {String: [String]}
    headers {String: [String]}
    body Body
  }

  // HTTP response structure used for HTTP ingress verbs.
  export data HttpResponse<Body, Error> {
    status Int
    headers {String: [String]}
    // Either "body" or "error" must be present, not both.
    body Body?
    error Error?
  }

  export data Empty {}

  // CatchRequest is a request structure for catch verbs.
  export data CatchRequest<Req> {
    request Req
    requestType Ref
    error String
  }
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
