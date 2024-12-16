package other

import (
	"context"
	"fmt"
	"time"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
	lib "github.com/block/ftl/go-runtime/schema/testdata"

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

type MyList []string

func (MyList) tag() {}

type MyMap map[string]string

func (MyMap) tag() {}

type MyString string

func (MyString) tag() {}

type MyStruct struct{}

func (MyStruct) tag() {}

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
	ExternalExternalType  another.External
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb export
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}

//ftl:typealias
type External lib.NonFTLType
