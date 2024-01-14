package internal

import (
	"archive/zip"
	"os"

	"github.com/TBD54566975/scaffolder"
)

// ScaffoldZip is a convenience function for scaffolding a zip archive with scaffolder.
func ScaffoldZip(source *zip.Reader, destination string, ctx any, options ...scaffolder.Option) error {
	tmpDir, err := os.MkdirTemp("", "scaffold-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	if err := UnzipDir(source, tmpDir); err != nil {
		return err
	}
	return scaffolder.Scaffold(tmpDir, destination, ctx, options...)
}
