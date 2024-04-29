package echo

import (
	"context"
	"ftl/time"
	"testing"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
	"github.com/alecthomas/assert/v2"

	stdtime "time"
)

func TestEcho(T *testing.T) {
	builder := ftltest.NewContextBuilder("echo")
	builder.AddConfig("default", "anonymous")
	builder.MockSourceVerb("time", "time", func(ctx context.Context) (any, error) {
		return time.TimeResponse{Time: stdtime.Date(2021, 9, 1, 0, 0, 0, 0, stdtime.UTC)}, nil
	})
	ctx, err := builder.Build()
	assert.NoError(T, err)

	resp, err := Echo(ctx, EchoRequest{Name: ftl.Some("world")})
	assert.NoError(T, err)
	assert.Equal(T, "Hello, world!!! It is 2021-09-01 00:00:00 +0000 UTC!", resp.Message)
}
