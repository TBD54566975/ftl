package generate

import (
	"io"
	"os"
	"path/filepath"
)

// File is a helper function for the generator functions in this package, to create a file and its parent directories, then call a generator function.
func File[T any](path string, importRoot string, generator func(io.Writer, T, string) error, parameter T) error {
	err := os.MkdirAll(filepath.Dir(path), 0o750)
	if err != nil {
		return err
	}
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close() //nolint:gosec
	return generator(w, parameter, importRoot)
}
