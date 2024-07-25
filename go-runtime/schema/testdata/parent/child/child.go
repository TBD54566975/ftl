package child

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type ChildStruct struct {
	Name      ftl.Option[ChildAlias]
	ValueEnum ChildValueEnum
	TypeEnum  ChildTypeEnum
}

type ChildAlias string

type Resp struct {
}

//ftl:verb
func ChildVerb(ctx context.Context) (Resp, error) {
	return Resp{}, nil
}

type ChildValueEnum int

const (
	A ChildValueEnum = iota
	B
	C
)

type ChildTypeEnum interface {
	tag()
}

type Scalar string

func (Scalar) tag() {}

type List []string

func (List) tag() {}
