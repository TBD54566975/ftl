package ftl

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/block/ftl/internal/log"
	. "github.com/block/ftl/internal/testutils/modulecontext"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/go-runtime/internal"
	"github.com/block/ftl/internal/deploymentcontext"
)

type intHandle int

func (s intHandle) Get(ctx context.Context) int { return int(s) }

func TestMapPanic(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx = internal.WithContext(context.Background(), internal.New(MakeDynamic(ctx, deploymentcontext.Empty("test"))))
	n := intHandle(1)
	once := Map(n, func(ctx context.Context, n int) (string, error) {
		return "", fmt.Errorf("test error %d", n)
	})
	assert.Panics(t, func() {
		once.Get(ctx)
	})
}

func TestMapGet(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx = internal.WithContext(context.Background(), internal.New(MakeDynamic(ctx, deploymentcontext.Empty("test"))))
	n := intHandle(1)
	once := Map(n, func(ctx context.Context, n int) (string, error) {
		return strconv.Itoa(n), nil
	})
	assert.Equal(t, once.Get(ctx), "1")
}
