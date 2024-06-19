//go:build !release

package internal

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
)

// ZipRelativeToCaller creates a temporary zip file from a path relative to the caller.
//
// This function will leak a file descriptor and thus can only be used in development.
func ZipRelativeToCaller(relativePath string) *zip.Reader {
	_, file, _, _ := runtime.Caller(1)
	dir := filepath.Join(filepath.Dir(file), relativePath)
	w, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.Remove(w.Name()) // This is okay because the zip.Reader will keep it open.
	if err != nil {
		panic(err)
	}

	err = ZipDir(dir, w.Name())
	if err != nil {
		panic(err)
	}

	info, err := w.Stat()
	if err != nil {
		panic(err)
	}
	_, err = w.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	zr, err := zip.NewReader(w, info.Size())
	if err != nil {
		panic(err)
	}
	return zr
}
