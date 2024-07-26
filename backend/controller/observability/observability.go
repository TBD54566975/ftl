package observability

import "fmt"

var (
	fsm *FSMMetrics
)

func InitControllerObservability() error {
	var err error

	fsm, err = InitFSMMetrics()

	if err != nil {
		return fmt.Errorf("could not initialize controller metrics: %w", err)
	}

	return nil
}

func FSM() *FSMMetrics {
	return fsm
}
