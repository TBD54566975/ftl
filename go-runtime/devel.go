//go:build !release

package goruntime

import (
	"archive/zip"

	"github.com/block/ftl/internal"
)

// Files is the FTL Go runtime scaffolding files.
func Files() *zip.Reader { return internal.ZipRelativeToCaller("scaffolding") }
