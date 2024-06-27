//go:build release

package builder

import (
	"archive/zip"
	"bytes"
	_ "embed"
)

//go:embed ../external-module-template.zip
var externalModuleTemplateBytes []byte

//go:embed ../build-template.zip
var buildTemplateBytes []byte

func externalModuleTemplateFiles() *zip.Reader {
	zr, err := zip.NewReader(bytes.NewReader(externalModuleTemplateBytes), int64(len(externalModuleTemplateBytes)))
	if err != nil {
		panic(err)
	}
	return zr
}

func buildTemplateFiles() *zip.Reader {
	zr, err := zip.NewReader(bytes.NewReader(buildTemplateBytes), int64(len(buildTemplateBytes)))
	if err != nil {
		panic(err)
	}
	return zr
}
