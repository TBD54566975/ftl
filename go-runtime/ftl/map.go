package ftl

import (
	"context"
	"sync"
)

type MapHandle[T, U any] struct {
	fn     func(context.Context, T) (U, error)
	getter Handle[T]
	out    U
	once   *sync.Once
}

func (sh *MapHandle[T, U]) Get(ctx context.Context) U {
	sh.once.Do(func() {
		t, err := sh.fn(ctx, sh.getter.Get(ctx))
		if err != nil {
			panic(err)
		}
		sh.out = t
	})
	return sh.out
}

// Map an FTL resource type to a new type.
func Map[T, U any](getter Handle[T], fn func(context.Context, T) (U, error)) MapHandle[T, U] {
	return MapHandle[T, U]{fn: fn, getter: getter, once: &sync.Once{}}
}
