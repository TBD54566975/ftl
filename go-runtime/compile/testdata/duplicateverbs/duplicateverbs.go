package duplicateverbs

import (
	"context"
	"time"
)

type TimeRequest struct {
	Name string
}
type TimeResponse struct {
	Time time.Time
}

// Time returns the current time.
//
//ftl:export
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}

//ftl:export
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}
