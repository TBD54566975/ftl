//ftl:module one
package one

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/compile/testdata/two"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:enum
type Color string

const (
	Red   Color = "Red"
	Blue  Color = "Blue"
	Green Color = "Green"
)

//ftl:enum
type ColorInt int

const (
	RedInt   ColorInt = 0
	BlueInt  ColorInt = 1
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

type Nested struct {
}

type Req struct {
	Int      int
	Int64    int64
	Float    float64
	String   string
	Slice    []string
	Map      map[string]string
	Nested   Nested
	Optional ftl.Option[Nested]
	Time     time.Time
	User     two.User `json:"u"`
	Bytes    []byte
	EnumRef  two.TwoEnum
}
type Resp struct{}

type Config struct {
	Field string
}

var configValue = ftl.Config[Config]("configValue")
var secretValue = ftl.Secret[string]("secretValue")

//ftl:verb
func Verb(ctx context.Context, req Req) (Resp, error) {
	return Resp{}, nil
}

const Yellow Color = "Yellow"

const YellowInt ColorInt = 3
