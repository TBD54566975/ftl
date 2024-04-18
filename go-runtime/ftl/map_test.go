package ftl

import (
	"context"
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

type intHandle int

func (s intHandle) Get(ctx context.Context) int { return int(s) }

func TestMapBaseCase(t *testing.T) {
	incrementer := 0

	h := intHandle(123)
	ctx := context.Background()
	once := Map(h, func(ctx context.Context, n int) (string, error) {
		incrementer++
		return fmt.Sprintf("handle: %d", n), nil
	})

	assert.Equal(t, once.Get(ctx), "handle: 123")
	assert.Equal(t, once.Get(ctx), "handle: 123")
	assert.Equal(t, incrementer, 1)
}

func TestMapPanic(t *testing.T) {
	ctx := context.Background()
	n := intHandle(1)
	once := Map(n, func(ctx context.Context, n int) (string, error) {
		return "", fmt.Errorf("test error %d", n)
	})
	assert.Panics(t, func() {
		once.Get(ctx)
	})
}
