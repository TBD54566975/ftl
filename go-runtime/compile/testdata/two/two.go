//ftl:module two
package two

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/sdk"
)

type User struct {
	Name string
}

type Payload[T any] struct {
	Body T
}

//ftl:verb
func Two(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return Payload[string]{}, nil
}

//ftl:verb
func CallsTwo(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return sdk.Call(ctx, Two, req)
}
