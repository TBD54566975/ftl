package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestTypeResolver(t *testing.T) {
	module, err := ParseModuleString("", `
		module test {
			data Request<T> {
				t T
			}
			verb test(test.Request<String>) Empty

			// This module has it's own definition of HttpRequest
			data HttpRequest {
			}
		}
	`)
	assert.NoError(t, err)
	otherModule, err := ParseModuleString("", `
		module other {
			data External {
			}
		}
	`)
	assert.NoError(t, err)
	scopes := NewScopes()
	err = scopes.Add(optional.None[*Module](), module.Name, module)
	assert.NoError(t, err)
	err = scopes.Add(optional.None[*Module](), otherModule.Name, otherModule)
	assert.NoError(t, err)

	// Resolving "HttpRequest" should return builtin.HttpRequest
	httpRequest := scopes.Resolve(Ref{Name: "HttpRequest"})
	assert.Equal(t, httpRequest.Module.MustGet().Name, "builtin")

	// Push a new scope for "test" module's decls
	scopes = scopes.Push()
	assert.NoError(t, scopes.AddModuleDecls(module))

	// Resolving "HttpRequest" should return test.HttpRequest now that we've pushed the new scope
	httpRequest = scopes.Resolve(Ref{Name: "HttpRequest"})
	assert.Equal(t, httpRequest.Module.MustGet().Name, "test")

	// Resolving "External" should fail
	external := scopes.Resolve(Ref{Name: "External"})
	assert.Equal(t, external, nil)

	// Resolve a builtin.
	actualInt, _ := ResolveAs[*Int](scopes, Ref{Name: "Int"})
	assert.NotZero(t, actualInt)
	assert.Equal(t, &Int{}, actualInt, assert.Exclude[Position]())

	// Resolve a generic data structure.
	actualData, _ := ResolveAs[*Data](scopes, Ref{Module: "test", Name: "Request"})
	assert.NotZero(t, actualData)
	expectedData := &Data{
		Name:           "Request",
		TypeParameters: []*TypeParameter{{Name: "T"}},
		Fields:         []*Field{{Name: "t", Type: &Ref{Name: "T"}}},
	}
	assert.Equal(t, expectedData, actualData, assert.Exclude[Position]())

	// Resolve a type parameter.
	scopes = scopes.PushScope(actualData.Scope())
	actualTP, _ := ResolveAs[*TypeParameter](scopes, Ref{Name: "T"})
	assert.Equal(t, actualTP, &TypeParameter{Name: "T"}, assert.Exclude[Position]())
}
