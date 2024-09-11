package time

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type TimeRequest struct{}
type TimeResponse struct {
	Time time.Time
}
type Testdb = ftl.PostgresDatabaseHandle

// Time returns the current time.
//
//ftl:verb export
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	//_, err := in(ctx, TimeRequest{})
	//if err != nil {
	//	return TimeResponse{}, err
	//}
	return TimeResponse{Time: time.Now()}, nil
}

//ftl:verb
func Internal(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}

//ftl:enum
type InternalEnum interface {
	internalEnum()
}

type Variant struct{}

func (Variant) internalEnum() {}
