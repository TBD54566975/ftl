package child

import (
	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type ChildStruct struct {
	Name ftl.Option[ChildAlias]
}

type ChildAlias string
