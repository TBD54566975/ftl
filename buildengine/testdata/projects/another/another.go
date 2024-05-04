package another

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

//ftl:enum export
type TypeEnum interface {
	tag()
}

type A int

func (A) tag() {}

type B string

func (B) tag() {}

//ftl:enum export
type SecondTypeEnum interface{ typeEnum() }

type One int

func (One) typeEnum() {}

type Two string

func (Two) typeEnum() {}

//ftl:data export
type TransitiveTypeEnum struct {
	TypeEnumRef SecondTypeEnum
}

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb export
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
