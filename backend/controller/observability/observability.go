package observability

import "fmt"

func InitControllerObservability() error {
	if err := InitFSMMetrics(); err != nil {
		return fmt.Errorf("could not initialize controller metrics: %w", err)
	}

	return nil
}
