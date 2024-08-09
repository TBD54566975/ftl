//go:build release

package goruntime

import (
	"archive/zip"
	"bytes"
	_ "embed"
)

//go:embed scaffolding.zip
var archive []byte

//go:embed external-module-template.zip
var externalModuleTemplate []byte

// Files is the FTL Java runtime scaffolding files.
func Files() *zip.Reader {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		panic(err)
	}
	return zr
}
