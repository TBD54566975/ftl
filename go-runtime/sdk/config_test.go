package sdk

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestConfig(t *testing.T) {
	t.Setenv("FTL_CONFIG_TESTING_TEST", `["one", "two", "three"]`)
	config := Config[[]string]("test")
	assert.Equal(t, []string{"one", "two", "three"}, config.Get())
}
