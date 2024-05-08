package ftl

import (
	"context"
)

type MapHandle[T, U any] struct {
	fn     func(context.Context, T) (U, error)
	getter Handle[T]
}

func (mh *MapHandle[T, U]) Get(ctx context.Context) U {
	t, err := mh.fn(ctx, mh.getter.Get(ctx))
	if err != nil {
		panic(err)
	}
	return t
}

// Map an FTL resource type to a new type.
func Map[T, U any](getter Handle[T], fn func(context.Context, T) (U, error)) MapHandle[T, U] {
	return MapHandle[T, U]{fn: fn, getter: getter}
}
