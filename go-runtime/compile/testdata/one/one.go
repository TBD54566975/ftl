package one

import (
	"context"
	"time"

	"ftl/two"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:export
type Color string

const (
	Red   Color = "Red"
	Blue  Color = "Blue"
	Green Color = "Green"
)

// Comments about ColorInt.
//
//ftl:export
type ColorInt int

const (
	// RedInt is a color.
	RedInt  ColorInt = 0
	BlueInt ColorInt = 1
	// GreenInt is also a color.
	GreenInt ColorInt = 2
)

//ftl:export
type SimpleIota int

const (
	Zero SimpleIota = iota
	One
	Two
)

//ftl:export
type IotaExpr int

const (
	First IotaExpr = iota*2 + 1
	Second
	Third
)

type Nested struct {
}

type Req struct {
	Int             int
	Int64           int64
	Float           float64
	String          string
	Slice           []string
	Map             map[string]string
	Nested          Nested
	Optional        ftl.Option[Nested]
	Time            time.Time
	User            two.User `json:"u"`
	Bytes           []byte
	LocalEnumRef    Color
	ExternalEnumRef two.TwoEnum
}
type Resp struct{}

type Config struct {
	Field string
}

var configValue = ftl.Config[Config]("configValue")
var secretValue = ftl.Secret[string]("secretValue")
var testDb = ftl.PostgresDatabase("testDb")

//ftl:export
func Verb(ctx context.Context, req Req) (Resp, error) {
	return Resp{}, nil
}

const Yellow Color = "Yellow"

const YellowInt ColorInt = 3

type SinkReq struct{}

//ftl:export
func Sink(ctx context.Context, req SinkReq) error {
	return nil
}

type SourceResp struct{}

//ftl:export
func Source(ctx context.Context) (SourceResp, error) {
	return SourceResp{}, nil
}

//ftl:export
func Nothing(ctx context.Context) error {
	return nil
}
