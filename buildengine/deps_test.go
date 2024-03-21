package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestExtractModuleDepsGo(t *testing.T) {
	deps, err := extractGoFTLImports("test", "testdata/projects/alpha")
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "other"}, deps)
}

func TestExtractModuleDepsKotlin(t *testing.T) {
	deps, err := extractKotlinFTLImports("test", "testdata/projects/alphakotlin")
	assert.NoError(t, err)
	assert.Equal(t, []string{"builtin", "other"}, deps)
}

func TestExtractLibraryDepsGo(t *testing.T) {
	deps, err := extractGoFTLImports("test", "testdata/projects/lib")
	assert.NoError(t, err)
	assert.Equal(t, []string{"alpha"}, deps)
}

func TestExtractLibraryDepsKotlin(t *testing.T) {
	deps, err := extractKotlinFTLImports("test", "testdata/projects/libkotlin")
	assert.NoError(t, err)
	assert.Equal(t, []string{"builtin", "echo"}, deps)
}
