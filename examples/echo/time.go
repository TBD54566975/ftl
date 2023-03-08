package echo

import (
	"context"
	"time"
)

type TimeRequest struct{}
type TimeResponse struct{ Time time.Time }

//ftl:verb
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}
