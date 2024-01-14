//go:build !release

package kotlinruntime

import (
	"archive/zip"

	"github.com/TBD54566975/ftl/internal"
)

// Files is the FTL Kotlin runtime scaffolding files.
func Files() *zip.Reader { return internal.ZipRelativeToCaller("scaffolding") }
