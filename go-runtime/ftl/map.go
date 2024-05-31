package ftl

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/internal"
)

type MapHandle[T, U any] struct {
	fn     func(context.Context, T) (U, error)
	getter Handle[T]
}

func (mh *MapHandle[T, U]) Get(ctx context.Context) U {
	out := internal.FromContext(ctx).CallMap(ctx, mh, func(ctx context.Context) (any, error) {
		return mh.fn(ctx, mh.getter.Get(ctx))
	})
	u, ok := out.(U)
	if !ok {
		panic(fmt.Sprintf("output object %v is not compatible with expected type %T", out, *new(U)))
	}
	return u
}

// Map an FTL resource type to a new type.
func Map[T, U any](getter Handle[T], fn func(context.Context, T) (U, error)) MapHandle[T, U] {
	return MapHandle[T, U]{
		fn:     fn,
		getter: getter,
	}
}
