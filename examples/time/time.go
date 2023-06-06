//ftl:module time
package time

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/sdk-go/observability"
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

	observability.Int64Counter(ctx, "called").Add(ctx, 1)

	_, span := observability.StartSpan(ctx, "amazing")
	defer span.End()

	return TimeResponse{Time: int(time.Now().Unix())}, nil
}
