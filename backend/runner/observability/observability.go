package observability

import (
	"errors"
	"fmt"
)

var (
	Runner     *RunnerMetrics
	Deployment *DeploymentMetrics
)

func init() {
	var errs error
	var err error

	Runner, err = initRunnerMetrics()
	errs = errors.Join(errs, err)
	Deployment, err = initDeploymentMetrics()
	errs = errors.Join(errs, err)

	if errs != nil {
		panic(fmt.Errorf("could not initialize runner metrics: %w", err))
	}
}

func wrapErr(signalName string, err error) error {
	return fmt.Errorf("failed to create %q signal: %w", signalName, err)
}
