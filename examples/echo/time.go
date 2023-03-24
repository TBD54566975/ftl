package echo

import (
	"context"
	"time"
)

type TimeRequest struct{}
type TimeResponse struct {
	Time int `json:"time"`
}

//ftl:verb
func GetTime(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: int(time.Now().Unix())}, nil
}
