package external

import (
	"context"

	lib "github.com/TBD54566975/ftl/go-runtime/schema/testdata"
)

type AliasedExternal lib.NonFTLType

//ftl:enum
type TypeEnum interface{ tag() }

type ExternalTypeVariant AliasedExternal

func (ExternalTypeVariant) tag() {}

//ftl:verb
func Echo(ctx context.Context, req AliasedExternal) (AliasedExternal, error) {
	return req, nil
}
