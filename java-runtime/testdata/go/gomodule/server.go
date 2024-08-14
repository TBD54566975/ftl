package gomodule

import (
	"context"
	"fmt"
	"time"
)

type TimeRequest struct {
}
type TimeResponse struct {
	Time time.Time
}

//ftl:verb export
func SourceVerb(ctx context.Context) (string, error) {
	return "Source Verb", nil
}

//ftl:verb export
func SinkVerb(ctx context.Context, req string) error {
	return nil
}

//ftl:verb export
func EmptyVerb(ctx context.Context) error {
	return nil
}

//ftl:verb export
func ErrorEmptyVerb(ctx context.Context) error {
	return fmt.Errorf("verb failed")
}

//ftl:verb export
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}
