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

func TestExtractDepsKotlin(t *testing.T) {
	deps, err := extractKotlinFTLImports("test", "testdata/modules/alphakotlin")
	assert.NoError(t, err)
	assert.Equal(t, []string{"builtin", "other"}, deps)
}
