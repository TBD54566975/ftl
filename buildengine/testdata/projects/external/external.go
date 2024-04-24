//ftl:module external
package external

import (
	"context"
	"time"
)

type ExternalRequest struct{}
type ExternalResponse struct {
	Month time.Month // external type should not be allowed
}

// External returns the current month as an external type.
//
//ftl:internal
func Time(ctx context.Context, req ExternalRequest) (ExternalResponse, error) {
	return ExternalResponse{Month: time.Now().Month()}, nil
}
