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

//ftl:verb
func Internal(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}

type SampleRequest struct {
	Name string
}

type SampleResponse struct {
	Message string
}

//ftl:verb
func Sample(ctx context.Context, req SampleRequest) (SampleResponse, error) {
	return SampleResponse{Message: "Hello, world!"}, nil
}
