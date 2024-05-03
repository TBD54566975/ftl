package wrapped

import (
	"context"
	"fmt"
	"ftl/time"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

// Wrapped module provides 2 verbs: Outer and Inner.
// Outer calls Inner and Inner calls time.Time.
// This module is useful to testing mocking verbs and setting up test config and secrets

var mySecret = ftl.Secret[string]("secret")
var myConfig = ftl.Config[string]("config")

type WrappedResponse struct {
	Output string `json:"output"`
	Secret string `json:"secret"`
	Config string `json:"config"`
}

//ftl:verb
func Outer(ctx context.Context) (WrappedResponse, error) {
	return ftl.CallSource(ctx, Inner)
}

//ftl:verb
func Inner(ctx context.Context) (WrappedResponse, error) {
	resp, err := ftl.Call(ctx, time.Time, time.TimeRequest{})
	if err != nil {
		return WrappedResponse{}, err
	}
	return WrappedResponse{
		Output: fmt.Sprintf("%v", resp.Time),
		Config: myConfig.Get(ctx),
		Secret: mySecret.Get(ctx),
	}, nil
}
