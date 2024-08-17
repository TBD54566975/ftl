package buildengine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestComputeRootDirs(t *testing.T) {
	dir := t.TempDir()
	libDir := filepath.Join(dir, "libs", "test")

	err := os.MkdirAll(libDir, 0700)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		baseDir  string
		patterns []string
		roots    []string
	}{
		{
			name:    "Simple Patterns",
			baseDir: filepath.Join(dir, "modules", "test"),
			patterns: []string{
				"**/*.go",
				"go.mod",
				"go.sum",
			},
			roots: []string{filepath.Join(dir, "modules", "test")},
		},
		{
			name:    "Upward Traversal",
			baseDir: filepath.Join(dir, "modules", "test"),
			patterns: []string{
				"../../libs/test/**/*.go",
			},
			roots: []string{
				filepath.Join(dir, "modules", "test"),
				filepath.Join(dir, "libs", "test"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roots := computeRootDirs(tt.baseDir, tt.patterns)
			assert.Compare(t, roots, tt.roots)
		})
	}
}
