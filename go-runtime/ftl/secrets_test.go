package ftl

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestSecret(t *testing.T) {
	t.Setenv("FTL_CONFIG", "testdata/ftl-project.toml")
	type C struct {
		One string
		Two string
	}
	config := Secret[C]("secret")
	assert.Equal(t, C{"one", "two"}, config.Get())
}
