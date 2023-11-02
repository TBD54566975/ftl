//go:build !release

package kotlinruntime

import (
	"github.com/TBD54566975/ftl/internal"
)

// Files is the FTL Kotlin runtime scaffolding files.
var Files = internal.ZipRelativeToCaller("scaffolding")
