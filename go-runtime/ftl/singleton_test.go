package ftl

import (
	"context"
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestSingletonBaseCase(t *testing.T) {
	incrementer := 0

	ctx := context.Background()
	once := Singleton[string](func(ctx context.Context) (string, error) {
		incrementer++
		return "only once", nil
	})

	assert.Equal(t, once.Get(ctx), "only once")
	assert.Equal(t, once.Get(ctx), "only once")
	assert.Equal(t, incrementer, 1)
}

func TestSingletonPanic(t *testing.T) {
	ctx := context.Background()
	once := Singleton[string](func(ctx context.Context) (string, error) {
		return "", fmt.Errorf("test error")
	})
	assert.Panics(t, func() {
		once.Get(ctx)
	})
}
