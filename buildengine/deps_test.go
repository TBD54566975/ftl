package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestExtractModuleDepsGo(t *testing.T) {
	deps, err := extractGoFTLImports("test", "testdata/alpha")
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "other"}, deps)
}

func TestExtractModuleDepsKotlin(t *testing.T) {
	deps, err := extractKotlinFTLImports("test", "testdata/alphakotlin")
	assert.NoError(t, err)
	assert.Equal(t, []string{"builtin", "other"}, deps)
}
