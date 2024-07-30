package observability

import (
	"errors"
	"fmt"
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
