//ftl:module time
package time

import (
	"context"
	"time"
)

type TimeRequest struct{}
type TimeResponse struct {
	Time int `json:"time"`
}

// Time returns the current time.
//
//ftl:verb
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: int(time.Now().Unix())}, nil
}
