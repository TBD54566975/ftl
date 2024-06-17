package child

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type ChildStruct struct {
	Name ftl.Option[ChildAlias]
}

type ChildAlias string

type Resp struct {
}

//ftl:verb
func ChildVerb(ctx context.Context) (Resp, error) {
	return Resp{}, nil
}
