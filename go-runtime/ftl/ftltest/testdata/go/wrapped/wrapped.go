package wrapped

import (
	"context"
	"fmt"
	"ftl/time"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

// Wrapped module provides 2 verbs: Outer and Inner.
// Outer calls Inner and Inner calls time.Time.
// This module is useful to testing mocking verbs and setting up test config and secrets

type Secret = ftl.Secret[string]
type Config = ftl.Config[string]

type WrappedResponse struct {
	Output string `json:"output"`
	Secret string `json:"secret"`
	Config string `json:"config"`
}

//ftl:verb
func Outer(ctx context.Context, inner InnerClient) (WrappedResponse, error) {
	return inner(ctx)
}

//ftl:verb
func Inner(ctx context.Context, tc time.TimeClient, myConfig Config, mySecret Secret) (WrappedResponse, error) {
	resp, err := tc(ctx, time.TimeRequest{})
	if err != nil {
		return WrappedResponse{}, err
	}
	return WrappedResponse{
		Output: fmt.Sprintf("%v", resp.Time),
		Config: myConfig.Get(ctx),
		Secret: mySecret.Get(ctx),
	}, nil
}
