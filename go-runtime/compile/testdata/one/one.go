//ftl:module one
package one

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/compile/testdata/two"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
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
}
type Resp struct{}

//ftl:verb
func Verb(ctx context.Context, req Req) (Resp, error) {
	return Resp{}, nil
}
