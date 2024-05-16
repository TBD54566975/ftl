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

//ftl:typealias
type MyBool bool

func (MyBool) tag() {}

//ftl:typealias
type MyBytes []byte

func (MyBytes) tag() {}

//ftl:typealias
type MyFloat float64

func (MyFloat) tag() {}

//ftl:typealias
type MyInt int

func (MyInt) tag() {}

//ftl:typealias
type MyTime time.Time

func (MyTime) tag() {}

//ftl:typealias
type List []string

func (List) tag() {}

//ftl:typealias
type Map map[string]string

func (Map) tag() {}

//ftl:typealias
type MyString string

func (MyString) tag() {}

type Struct struct{}

func (Struct) tag() {}

//ftl:typealias
type MyOption ftl.Option[string]

func (MyOption) tag() {}

//ftl:typealias
type MyUnit ftl.Unit

func (MyUnit) tag() {}

//ftl:enum
type SecondTypeEnum interface {
	tag2()
}

//ftl:typealias
type A string

func (A) tag2() {}

//ftl:typealias
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
