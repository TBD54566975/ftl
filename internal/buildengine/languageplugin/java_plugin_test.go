package languageplugin

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestExtractModuleDepsKotlin(t *testing.T) {
	deps, err := extractKotlinFTLImports("test", "testdata/alphakotlin")
	assert.NoError(t, err)
	assert.Equal(t, []string{"builtin", "other"}, deps)
}
