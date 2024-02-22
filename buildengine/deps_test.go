package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestExtractDepsGo(t *testing.T) {
	deps, err := extractGoFTLImports("test", "testdata/modules/alpha")
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "other"}, deps)
}
