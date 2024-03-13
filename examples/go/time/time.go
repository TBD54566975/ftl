//ftl:module time
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
//ftl:verb
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}

//ftl:verb
func Sink(ctx context.Context, req TimeRequest) error {
	return nil
}

//ftl:verb
func Source(ctx context.Context) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}

//ftl:verb
func Empty(ctx context.Context) error {
	return nil
}
