package wrapped

import (
	"context"
	"ftl/time"
	"testing"
	stdtime "time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
	"github.com/alecthomas/assert/v2"
)

func TestWrapped(t *testing.T) {
	allOptions := []func(*ftltest.Options) error{
		ftltest.WithConfig(myConfig, "helloworld"),
		ftltest.WithSecret(mySecret, "shhhhh"),
		ftltest.WithCallsAllowedWithinModule(),
		ftltest.WhenVerb(time.Time, func(ctx context.Context, req time.TimeRequest) (time.TimeResponse, error) {
			return time.TimeResponse{Time: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)}, nil
		}),
	}

	for _, tt := range []struct {
		name          string
		options       []func(*ftltest.Options) error
		expectedError ftl.Option[string]
	}{
		{
			name:          "OnlyConfigAndSecret",
			options:       allOptions[:2],
			expectedError: ftl.Some("wrapped.inner: no mock found: provide a mock with ftltest.WhenVerb(Inner, ...) or enable all calls within the module with ftltest.WithCallsAllowedWithinModule()"),
		},
		{
			name:          "AllowCallsWithinModule",
			options:       allOptions[:3],
			expectedError: ftl.Some("wrapped.inner: time.time: no mock found: provide a mock with ftltest.WhenVerb(time.Time, ...)"),
		},
		{
			name:    "WithExternalVerbMock",
			options: allOptions[:4],
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ftltest.Context(
				tt.options...,
			)
			myConfig.Get(ctx)
			resp, err := Outer(ctx)

			if expected, ok := tt.expectedError.Get(); ok {
				assert.EqualError(t, err, expected)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, "2024-01-01 00:00:00 +0000 UTC", resp.Output)
			assert.Equal(t, "helloworld", resp.Config)
			assert.Equal(t, "shhhhh", resp.Secret)
		})
	}
}
