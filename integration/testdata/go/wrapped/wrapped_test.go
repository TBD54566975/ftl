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
	for _, tt := range []struct {
		name          string
		options       []ftltest.Option
		configValue   string
		secretValue   string
		expectedError ftl.Option[string]
	}{
		{
			name: "OnlyConfigAndSecret",
			options: []ftltest.Option{
				ftltest.WithConfig(myConfig, "helloworld"),
				ftltest.WithSecret(mySecret, "shhhhh"),
			},
			configValue:   "helloworld",
			secretValue:   "shhhhh",
			expectedError: ftl.Some("wrapped.inner: no mock found: provide a mock with ftltest.WhenVerb(Inner, ...) or enable all calls within the module with ftltest.WithCallsAllowedWithinModule()"),
		},
		{
			name: "AllowCallsWithinModule",
			options: []ftltest.Option{
				ftltest.WithConfig(myConfig, "helloworld"),
				ftltest.WithSecret(mySecret, "shhhhh"),
				ftltest.WithCallsAllowedWithinModule(),
			},
			configValue:   "helloworld",
			secretValue:   "shhhhh",
			expectedError: ftl.Some("wrapped.inner: time.time: no mock found: provide a mock with ftltest.WhenVerb(time.Time, ...)"),
		},
		{
			name: "WithExternalVerbMock",
			options: []ftltest.Option{
				ftltest.WithConfig(myConfig, "helloworld"),
				ftltest.WithSecret(mySecret, "shhhhh"),
				ftltest.WithCallsAllowedWithinModule(),
				ftltest.WhenVerb(time.Time, func(ctx context.Context, req time.TimeRequest) (time.TimeResponse, error) {
					return time.TimeResponse{Time: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)}, nil
				}),
			},
			configValue: "helloworld",
			secretValue: "shhhhh",
		},
		{
			name: "WithProjectToml",
			options: []ftltest.Option{
				ftltest.WithProjectFile("ftl-project-test.toml"),
				ftltest.WithCallsAllowedWithinModule(),
				ftltest.WhenVerb(time.Time, func(ctx context.Context, req time.TimeRequest) (time.TimeResponse, error) {
					return time.TimeResponse{Time: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)}, nil
				}),
			},
			configValue: "bar",
			secretValue: "bar",
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
			assert.Equal(t, tt.configValue, resp.Config)
			assert.Equal(t, tt.secretValue, resp.Secret)
		})
	}
}
