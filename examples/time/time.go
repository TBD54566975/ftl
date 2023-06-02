//ftl:module time
package time

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/observability"
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
	counter, err := observability.MeterProviderFromContext(ctx).Meter("time.time").Int64Counter("called")
	if err != nil {
		return TimeResponse{}, err
	}
	counter.Add(ctx, 1)
	return TimeResponse{Time: int(time.Now().Unix())}, nil
}
