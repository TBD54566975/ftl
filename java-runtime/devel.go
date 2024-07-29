//go:build !release

package kotlinruntime

import (
	"archive/zip"

	"github.com/TBD54566975/ftl/internal"
)

// Files is the FTL Java runtime scaffolding files.
func Files() *zip.Reader { return internal.ZipRelativeToCaller("scaffolding") }

// ExternalModuleTemplates are templates for scaffolding external modules in the FTL Java runtime.
func ExternalModuleTemplates() *zip.Reader {
	return internal.ZipRelativeToCaller("external-module-template")
}
