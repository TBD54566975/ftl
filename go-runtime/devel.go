//go:build !release

package goruntime

import (
	"github.com/TBD54566975/ftl/internal"
)

// Files is the FTL Go runtime scaffolding files.
var Files = internal.ZipRelativeToCaller("scaffolding")
