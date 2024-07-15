package one

import (
	"context"
	"errors"
	"time"

	"ftl/builtin"
	"ftl/two"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:enum export
type Color string

const (
	Red   Color = "Red"
	Blue  Color = "Blue"
	Green Color = "Green"
)

// Comments about ColorInt.
//
//ftl:enum
type ColorInt int

const (
	// RedInt is a color.
	RedInt  ColorInt = 0
	BlueInt ColorInt = 1
	// GreenInt is also a color.
	GreenInt ColorInt = 2
)

//ftl:enum
type SimpleIota int

const (
	Zero SimpleIota = iota
	One
	Two
)

//ftl:enum
type IotaExpr int

const (
	First IotaExpr = iota*2 + 1
	Second
	Third
)

//ftl:enum
type BlobOrList interface{ blobOrList() }

type Blob string

func (Blob) blobOrList() {}

type List []string

func (List) blobOrList() {}

type Nested struct {
}

//ftl:enum
type TypeEnum interface {
	tag()
}

type Option ftl.Option[string]

func (Option) tag() {}

type InlineStruct struct{}

func (InlineStruct) tag() {}

type AliasedStruct UnderlyingStruct

func (AliasedStruct) tag() {}

type UnderlyingStruct struct{}

type ValueEnum ColorInt

func (ValueEnum) tag() {}

//ftl:enum
type PrivateEnum interface{ privateEnum() }

//ftl:data export
type ExportedStruct struct{}

func (ExportedStruct) privateEnum() {}

//ftl:data
type PrivateStruct struct{}

func (PrivateStruct) privateEnum() {}

type WithoutDirectiveStruct struct{}

func (WithoutDirectiveStruct) privateEnum() {}

type Req struct {
	Int                  int
	Float                float64
	String               string
	Slice                []string
	Map                  map[string]string
	Nested               Nested
	Optional             ftl.Option[Nested]
	Time                 time.Time
	User                 two.User `json:"u"`
	Bytes                []byte
	LocalValueEnumRef    Color
	LocalTypeEnumRef     BlobOrList
	ExternalValueEnumRef two.TwoEnum
	ExternalTypeEnumRef  two.TypeEnum
}
type Resp struct{}

type Config struct {
	Field string
}

//ftl:data export
type ExportedData struct {
	Field string
}

var configValue = ftl.Config[Config]("configValue")
var secretValue = ftl.Secret[string]("secretValue")
var testDb = ftl.PostgresDatabase("testDb")

//ftl:verb
func Verb(ctx context.Context, req Req) (Resp, error) {
	return Resp{}, nil
}

const Yellow Color = "Yellow"

const YellowInt ColorInt = 3

type SinkReq struct{}

//ftl:verb
func Sink(ctx context.Context, req SinkReq) error {
	return nil
}

type SourceResp struct{}

//ftl:verb
func Source(ctx context.Context) (SourceResp, error) {
	return SourceResp{}, nil
}

//ftl:verb export
func Nothing(ctx context.Context) error {
	return nil
}

//ftl:ingress http GET /get
func Http(ctx context.Context, req builtin.HttpRequest[Req]) (builtin.HttpResponse[Resp, ftl.Unit], error) {
	return builtin.HttpResponse[Resp, ftl.Unit]{}, nil
}

//ftl:data
type DataWithType[T any] struct {
	Value T
}

type NonFTLInterface interface {
	NonFTLInterface()
}

type NonFTLStruct struct {
	Name string
}

func (NonFTLStruct) NonFTLInterface() {}

//ftl:verb
func StringToTime(ctx context.Context, input string) (time.Time, error) {
	return time.Time{}, errors.New("not implemented")
}

//ftl:verb
func BatchStringToTime(ctx context.Context, input []string) ([]time.Time, error) {
	return nil, errors.New("not implemented")
}
