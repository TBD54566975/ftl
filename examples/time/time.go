//ftl:module time
package time

import (
	"context"
	"time"
)

type TimeRequest struct{}
type TimeResponse struct {
	Time time.Time `json:"time"`
}

// Time returns the current time.
//
//ftl:verb
//ftl:ingress GET /time
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{
		Time: time.Now(),
	}, nil
}
