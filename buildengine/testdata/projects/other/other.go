package other

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.

	"ftl/another"
)

//ftl:enum
type TypeEnum interface {
	tag()
}

type MyBool bool

func (MyBool) tag() {}

type MyBytes []byte

func (MyBytes) tag() {}

type MyFloat float64

func (MyFloat) tag() {}

type MyInt int

func (MyInt) tag() {}

type MyTime time.Time

func (MyTime) tag() {}

type List []string

func (List) tag() {}

type Map map[string]string

func (Map) tag() {}

type MyString string

func (MyString) tag() {}

type Struct struct{}

func (Struct) tag() {}

type MyOption ftl.Option[string]

func (MyOption) tag() {}

type MyUnit ftl.Unit

func (MyUnit) tag() {}

//ftl:enum
type SecondTypeEnum interface {
	tag2()
}

type A string

func (A) tag2() {}

type B EchoRequest

func (B) tag2() {}

type EchoRequest struct {
	Name                  ftl.Option[string] `json:"name"`
	ExternalSumType       another.TypeEnum
	ExternalNestedSumType another.TransitiveTypeEnum
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb export
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
