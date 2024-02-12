package ftl

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestConfig(t *testing.T) {
	type C struct {
		One string
		Two string
	}
	t.Setenv("FTL_CONFIG_TESTING_TEST", `{"one": "one", "two": "two"}`)
	config := Config[C]("test")
	assert.Equal(t, C{"one", "two"}, config.Get())
}
