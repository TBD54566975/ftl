//ftl:module time
package time

import (
	"context"
	"time"

	"ftl/builtin"
)

type TimeRequest struct{}
type TimeResponse struct {
	Time time.Time `json:"time"`
}

// Time returns the current time.
//
//ftl:verb
//ftl:ingress GET /time
func Time(ctx context.Context, req builtin.HttpRequest[TimeRequest]) (builtin.HttpResponse[TimeResponse], error) {
	return builtin.HttpResponse[TimeResponse]{
		Status:  200,
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body:    TimeResponse{Time: time.Now()},
	}, nil
}
