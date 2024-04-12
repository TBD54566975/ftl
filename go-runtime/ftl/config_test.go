package ftl

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestConfig(t *testing.T) {
	t.Setenv("FTL_CONFIG", "testdata/ftl-project.toml")
	type C struct {
		One string
		Two string
	}
	config := Config[C]("test")
	assert.Equal(t, C{"one", "two"}, config.Get())
}
