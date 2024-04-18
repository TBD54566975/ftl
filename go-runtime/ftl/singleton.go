package ftl

import (
	"context"
	"sync"
)

type InputLambda[T any] func(context.Context) (T, error)

type SingletonConstructor[T any] struct {
	Fn   InputLambda[T]
	Out  T
	Once sync.Once
}

func (sf *SingletonConstructor[T]) Get(ctx context.Context) T {
	sf.Once.Do(func() {
		t, err := sf.Fn(ctx)
		if err != nil {
			panic(err)
		}
		sf.Out = t
	})
	return sf.Out
}

func Singleton[T any](fn InputLambda[T]) SingletonConstructor[T] {
    return SingletonConstructor[T]{Fn: fn}
}
