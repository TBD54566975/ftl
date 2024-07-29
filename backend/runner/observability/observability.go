package observability

import (
	"fmt"
)

var (
	RUNNER *FSMMetrics
)

func init() {
	var err error

	FSM, err = initFSMMetrics()

	if err != nil {
		panic(fmt.Errorf("could not initialize controller metrics: %w\n", err))
	}
}
