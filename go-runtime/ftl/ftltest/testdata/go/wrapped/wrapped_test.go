package wrapped

import (
	"context"
	"ftl/time"
	"path/filepath"
	"testing"
	stdtime "time"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/ftl/ftltest"
)

func TestWrappedWithConfigEnvar(t *testing.T) {
	absProjectPath1, err := filepath.Abs("ftl-project-test-1.toml")
	assert.NoError(t, err)
	t.Setenv("FTL_CONFIG", absProjectPath1)

	for _, tt := range []struct {
		name          string
		options       []ftltest.Option
		configValue   string
		secretValue   string
		expectedError ftl.Option[string]
	}{
		{
			name: "WithProjectTomlFromEnvar",
			options: []ftltest.Option{
				ftltest.WithDefaultProjectFile(),
				ftltest.WithCallsAllowedWithinModule(),
				ftltest.WhenVerb[time.TimeClient, time.TimeRequest, time.TimeResponse](func(ctx context.Context, req time.TimeRequest) (time.TimeResponse, error) {
					return time.TimeResponse{Time: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)}, nil
				}),
			},
			configValue: "foobar",
			secretValue: "foobar",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ftltest.Context(
				tt.options...,
			)
			// TODO: need a GetConfigValue test helper
			// myConfig.Get(ctx)
			ftl.Config[string]{Ref: reflection.Ref{Module: "wrapped", Name: "config"}}.Get(ctx)

			resp, err := ftltest.CallSource[OuterClient, WrappedResponse](ctx)

			if expected, ok := tt.expectedError.Get(); ok {
				assert.EqualError(t, err, expected)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, WrappedResponse{
				Output: "2024-01-01 00:00:00 +0000 UTC",
				Config: tt.configValue,
				Secret: tt.secretValue,
			}, resp)
		})
	}
}

// TODO: add test helpers for setting configs and secrets
//
// func TestWrapped(t *testing.T) {
// 	for _, tt := range []struct {
// 		name          string
// 		options       []ftltest.Option
// 		configValue   string
// 		secretValue   string
// 		expectedError ftl.Option[string]
// 	}{
// 		{
// 			name: "OnlyConfigAndSecret",
// 			options: []ftltest.Option{
// 				ftltest.WithConfig(myConfig, "helloworld"),
// 				ftltest.WithSecret(mySecret, "shhhhh"),
// 			},
// 			configValue:   "helloworld",
// 			secretValue:   "shhhhh",
// 			expectedError: ftl.Some("test harness failed to call verb wrapped.outer: wrapped.inner: no mock found: provide a mock with ftltest.WhenVerb(Inner, ...) or enable all calls within the module with ftltest.WithCallsAllowedWithinModule()"),
// 		},
// 		{
// 			name: "AllowCallsWithinModule",
// 			options: []ftltest.Option{
// 				ftltest.WithConfig(myConfig, "helloworld"),
// 				ftltest.WithSecret(mySecret, "shhhhh"),
// 				ftltest.WithCallsAllowedWithinModule(),
// 			},
// 			configValue:   "helloworld",
// 			secretValue:   "shhhhh",
// 			expectedError: ftl.Some("test harness failed to call verb wrapped.outer: wrapped.inner: time.time: no mock found: provide a mock with ftltest.WhenVerb(time.Time, ...)"),
// 		},
// 		{
// 			name: "WithExternalVerbMock",
// 			options: []ftltest.Option{
// 				ftltest.WithConfig(myConfig, "helloworld"),
// 				ftltest.WithSecret(mySecret, "shhhhh"),
// 				ftltest.WithCallsAllowedWithinModule(),
// 				ftltest.WhenVerb[time.TimeClient, time.TimeRequest, time.TimeResponse](func(ctx context.Context, req time.TimeRequest) (time.TimeResponse, error) {
// 					return time.TimeResponse{Time: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)}, nil
// 				}),
// 			},
// 			configValue: "helloworld",
// 			secretValue: "shhhhh",
// 		},
// 		{
// 			name: "WithProjectTomlSpecified",
// 			options: []ftltest.Option{
// 				ftltest.WithProjectFile("ftl-project-test-1.toml"),
// 				ftltest.WithCallsAllowedWithinModule(),
// 				ftltest.WhenVerb[time.TimeClient, time.TimeRequest, time.TimeResponse](func(ctx context.Context, req time.TimeRequest) (time.TimeResponse, error) {
// 					return time.TimeResponse{Time: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)}, nil
// 				}),
// 			},
// 			configValue: "foobar",
// 			secretValue: "foobar",
// 		}, {
// 			name: "WithProjectTomlFromRoot",
// 			options: []ftltest.Option{
// 				ftltest.WithDefaultProjectFile(),
// 				ftltest.WithCallsAllowedWithinModule(),
// 				ftltest.WhenVerb[time.TimeClient, time.TimeRequest, time.TimeResponse](func(ctx context.Context, req time.TimeRequest) (time.TimeResponse, error) {
// 					return time.TimeResponse{Time: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)}, nil
// 				}),
// 			},
// 			configValue: "bazbaz",
// 			secretValue: "bazbaz",
// 		},
// 	} {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx := ftltest.Context(
// 				tt.options...,
// 			)
// 			myConfig.Get(ctx)
// 			resp, err := ftltest.CallSource[OuterClient, WrappedResponse](ctx)

// 			if expected, ok := tt.expectedError.Get(); ok {
// 				assert.EqualError(t, err, expected)
// 				return
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			assert.Equal(t, "2024-01-01 00:00:00 +0000 UTC", resp.Output)
// 			assert.Equal(t, tt.configValue, resp.Config)
// 			assert.Equal(t, tt.secretValue, resp.Secret)
// 		})
// 	}
// }
