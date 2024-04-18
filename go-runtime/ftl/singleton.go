package ftl

import (
	"context"
	"sync"
)

type SingletonHandle[T any] struct {
	fn   func(context.Context) (T, error)
	out  T
	once *sync.Once
}

func (sh *SingletonHandle[T]) Get(ctx context.Context) T {
	sh.once.Do(func() {
		t, err := sh.fn(ctx)
		if err != nil {
			panic(err)
		}
		sh.out = t
	})
	return sh.out
}

func Singleton[T any](fn func(context.Context) (T, error)) SingletonHandle[T] {
	return SingletonHandle[T]{fn: fn, once: &sync.Once{}}
}
