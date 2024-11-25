package schema

import (
	"bytes"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestRedact(t *testing.T) {
	module := &Module{
		Decls: []Decl{
			&Database{
				Runtime: &DSNDatabaseRuntime{
					DSN: "sensitive",
				},
			},
		},
	}
	redacted := Redact(module)
	assert.NotEqual(t, module, redacted)

	data, err := ModuleToBytes(module)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(data, []byte("sensitive")), "data should contain sensitive information")

	data, err = ModuleToBytes(redacted)
	assert.NoError(t, err)
	assert.False(t, bytes.Contains(data, []byte("sensitive")), "data should not contain sensitive information")
}
