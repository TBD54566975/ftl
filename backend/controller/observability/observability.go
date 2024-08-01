package observability

import (
	"errors"
	"fmt"
	"time"
)

var (
	AsyncCalls *AsyncCallMetrics
	FSM        *FSMMetrics
	PubSub     *PubSubMetrics
)

func init() {
	var errs error
	var err error

	AsyncCalls, err = initAsyncCallMetrics()
	errs = errors.Join(errs, err)
	FSM, err = initFSMMetrics()
	errs = errors.Join(errs, err)
	PubSub, err = initPubSubMetrics()
	errs = errors.Join(errs, err)

	if err != nil {
		panic(fmt.Errorf("could not initialize controller metrics: %w", errs))
	}
}

func timeSinceMS(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
