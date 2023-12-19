//go:build !release

package goruntime

import (
	"archive/zip"

	"github.com/TBD54566975/ftl/internal"
)

// Files is the FTL Go runtime scaffolding files.
func Files() *zip.Reader { return internal.ZipRelativeToCaller("scaffolding") }
