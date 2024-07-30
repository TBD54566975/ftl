package observability

import (
	"errors"
	"fmt"
)

var (
	FSM    *FSMMetrics
	PubSub *PubSubMetrics
)

func init() {
	var errs error
	var err error

	FSM, err = initFSMMetrics()
	errs = errors.Join(errs, err)
	PubSub, err = initPubSubMetrics()
	errs = errors.Join(errs, err)

	if err != nil {
		panic(fmt.Errorf("could not initialize controller metrics: %w", errs))
	}
}
