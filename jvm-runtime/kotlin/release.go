//go:build release

package kotlin

import (
	"archive/zip"
	"bytes"
	_ "embed"
)

//go:embed scaffolding.zip
var archive []byte

// Files is the FTL Go runtime scaffolding files.
func Files() *zip.Reader {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		panic(err)
	}
	return zr
}
