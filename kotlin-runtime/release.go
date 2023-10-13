//go:build release

package goruntime

import (
	"archive/zip"
	"bytes"
	_ "embed"
)

//go:embed scaffolding.zip
var archive []byte

// Files is the FTL Kotlin runtime scaffolding files.
//
// scaffolding.zip can be generated by running `bit kotlin-runtime/scaffolding.zip`
// or indirectly via `bit build/release/ftl`.
var Files = func() *zip.Reader {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		panic(err)
	}
	return zr
}()
