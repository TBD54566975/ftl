package ftl

import (
	"bytes"
	"context"
	"sync"
)

type MapHandle[T, U any] struct {
	fn     func(context.Context, T) (U, error)
	getter Handle[T]
	inHash []byte
	out    U
	once   *sync.Once
}

func (mh *MapHandle[T, U]) Get(ctx context.Context) U {
	//TODO: this is no longer threadsafe
	var latestHash []byte
	if h, ok := mh.getter.(HashableHandle[T]); ok {
		latestHash = h.Hash(ctx)
	} else {
		latestHash = []byte{}
	}

	if !bytes.Equal(mh.inHash, latestHash) {
		mh.once = &sync.Once{}
		mh.inHash = latestHash
	}
	mh.once.Do(func() {
		t, err := mh.fn(ctx, mh.getter.Get(ctx))
		if err != nil {
			panic(err)
		}
		mh.out = t
	})
	return mh.out
}

// Map an FTL resource type to a new type.
func Map[T, U any](getter Handle[T], fn func(context.Context, T) (U, error)) MapHandle[T, U] {
	return MapHandle[T, U]{fn: fn, getter: getter, once: &sync.Once{}, inHash: []byte{}}
}
