//ftl:module time
package time

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
)

type TimeRequest struct{}
type TimeResponse struct {
	Time int `json:"time"`
	Date int `json:"date"`
}

// Time returns the current time.
//
//ftl:verb
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	counter, err := otel.Meter("time.time").Int64Counter("called")
	if err != nil {
		return TimeResponse{}, err
	}
	counter.Add(ctx, 1)

	tracer := otel.Tracer("time.time")
	_, span := tracer.Start(ctx, "time.time")
	defer span.End()

	return TimeResponse{Time: int(time.Now().Unix())}, nil
}
