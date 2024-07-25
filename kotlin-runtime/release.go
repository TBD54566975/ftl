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

// Files is the FTL Kotlin runtime scaffolding files.
func Files() *zip.Reader {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		panic(err)
	}
	return zr
}

// ExternalModuleTemplates are templates for scaffolding external modules in the FTL Kotlin runtime.
func ExternalModuleTemplates() *zip.Reader {
	zr, err := zip.NewReader(bytes.NewReader(externalModuleTemplate), int64(len(externalModuleTemplate)))
	if err != nil {
		panic(err)
	}
	return zr
}
