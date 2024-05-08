package ftl

import (
	"context"
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

type intHandle int

func (s intHandle) Get(ctx context.Context) int { return int(s) }

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
