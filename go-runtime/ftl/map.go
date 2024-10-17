package ftl

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/internal"
)

type MapHandle[T, U any] struct {
	fn     func(context.Context, T) (U, error)
	handle Handle[T]
}

// Get the mapped value.
func (mh *MapHandle[T, U]) Get(ctx context.Context) U {
	value := mh.handle.Get(ctx)
	out := internal.FromContext(ctx).CallMap(ctx, mh, value, func(ctx context.Context) (any, error) {
		return mh.fn(ctx, value)
	})
	u, ok := out.(U)
	if !ok {
		panic(fmt.Sprintf("output object %v is not compatible with expected type %T", out, *new(U)))
	}
	return u
}

// Map an FTL resource type to a new type.
func Map[T, U any](getter Handle[T], fn func(context.Context, T) (U, error)) *MapHandle[T, U] {
	return &MapHandle[T, U]{
		fn:     fn,
		handle: getter,
	}
}

// ResourceMapper is an interface that can be implemented to map a resource of type `From` to a new type `To`.
//
// A ResourceMapper struct should embed the Handle type that it is mapping from and define a `Map` method that
// returns the new type.
//
// e.g.
//
//	 type DBMapper struct {
//	   ftl.DatabaseHandle[MyConfig]
//	 }
//
//		func (DBMapper) Map(ctx context.Context, conn *sql.DB) (NewType, error) {
//		  ...
//		}
type ResourceMapper[From, To any] interface {
	Handle[From]
	Map(ctx context.Context, f From) (To, error)
}

// MappedHandle can be used to provide the mapped value of a resource in a verb signature.
//
// e.g.
// //ftl:verb
//
//	func MyVerb(ctx context.Context, req Request, handle ftl.MappedHandle[DBMapper, *sql.DB, NewType]) (Response, error) {
//	  db := handle.Get(ctx)
//	  ...
//	}
type MappedHandle[R ResourceMapper[From, To], From, To any] struct {
	mapper R
}

func (mh *MappedHandle[H, From, To]) Get(ctx context.Context) To {
	value := mh.mapper.Get(ctx)
	out := internal.FromContext(ctx).CallMap(ctx, mh, value, func(ctx context.Context) (any, error) {
		return mh.mapper.Map(ctx, value)
	})
	u, ok := out.(To)
	if !ok {
		panic(fmt.Sprintf("output object %v is not compatible with expected type %T", out, *new(To)))
	}
	return u
}

func NewMappedHandle[H ResourceMapper[From, To], From, To any](mapper H) MappedHandle[H, From, To] {
	return MappedHandle[H, From, To]{mapper}
}
