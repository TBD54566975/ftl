package observability

import (
	"fmt"
)

var (
	Runner *RunnerMetrics
)

func init() {
	var err error

	Runner, err = initRunnerMetrics()

	if err != nil {
		panic(fmt.Errorf("could not initialize runner metrics: %w", err))
	}
}
