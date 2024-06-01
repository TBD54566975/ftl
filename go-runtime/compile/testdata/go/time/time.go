package time

import (
	"context"
	"time"
)

type TimeRequest struct{}
type TimeResponse struct {
	Time time.Time
}

// Time returns the current time.
//
//ftl:verb export
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}
