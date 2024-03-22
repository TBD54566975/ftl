package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTypeResolver(t *testing.T) {
	module, err := ParseModuleString("", `
		module test {
			data Request<T> {
				t T
			}
			verb test(Request<String>) Empty
		}
	`)
	assert.NoError(t, err)
	scopes := NewScopes()
	scopes = scopes.PushScope(module.Scope())

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
