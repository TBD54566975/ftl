package ftl

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestSecret(t *testing.T) {
	type C struct {
		One string
		Two string
	}
	t.Setenv("FTL_SECRET_TESTING_TEST", `{"one": "one", "two": "two"}`)
	config := Secret[C]("test")
	assert.Equal(t, C{"one", "two"}, config.Get())
}
