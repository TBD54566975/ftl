package ftl

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestSingletonBaseCase(t *testing.T) {
	incrementer := 0

	ctx := context.Background()
	once := Singleton[string](func(ctx context.Context) (string, error) {
		incrementer += 1
		return "only once", nil
	})

	assert.Equal(t, once.Get(ctx), "only once")
	assert.Equal(t, once.Get(ctx), "only once")
	assert.Equal(t, incrementer, 1)
}
