package mapper

import (
	"context"
	"strconv"
	"sync/atomic"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var underlyingCounter atomic.Int64
var mapCounter atomic.Int64

type value struct {
	value atomic.Int64
}

func (v *value) Get(ctx context.Context) int {
	underlyingCounter.Add(1)
	return int(v.value.Load())
}

var underlyingValue = new(value)
var mappedValue = ftl.Map(underlyingValue, func(ctx context.Context, in int) (string, error) {
	mapCounter.Add(1)
	return strconv.Itoa(in), nil

})

type Output struct {
	UnderlyingCounter int
	MapCounter        int
	Mapped            string
}

//ftl:verb
func Get(ctx context.Context) (Output, error) {
	mapped := mappedValue.Get(ctx)
	return Output{
		UnderlyingCounter: int(underlyingCounter.Load()),
		MapCounter:        int(mapCounter.Load()),
		Mapped:            mapped,
	}, nil
}

//ftl:verb
func Inc(ctx context.Context) error {
	underlyingValue.value.Add(1)
	return nil
}
