//go:build !release

package internal

import (
	"os"
	"path/filepath"
)

func FTLSourceRoot() string {
	return filepath.Clean(filepath.Join(filepath.Dir(os.Args[0]), "..", ".."))
}
