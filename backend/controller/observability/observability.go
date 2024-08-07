package observability

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	AsyncCalls *AsyncCallMetrics
	Calls      *CallMetrics
	Deployment *DeploymentMetrics
	FSM        *FSMMetrics
	PubSub     *PubSubMetrics
	Cron       *CronMetrics
)

func init() {
	var errs error
	var err error

	AsyncCalls, err = initAsyncCallMetrics()
	errs = errors.Join(errs, err)
	Calls, err = initCallMetrics()
	errs = errors.Join(errs, err)
	Deployment, err = initDeploymentMetrics()
	errs = errors.Join(errs, err)
	FSM, err = initFSMMetrics()
	errs = errors.Join(errs, err)
	PubSub, err = initPubSubMetrics()
	errs = errors.Join(errs, err)
	Cron, err = initCronMetrics()
	errs = errors.Join(errs, err)

	if err != nil {
		panic(fmt.Errorf("could not initialize controller metrics: %w", errs))
	}
}

func wrapErr(signalName string, err error) error {
	return fmt.Errorf("failed to create %q signal: %w", signalName, err)
}

func timeSinceMS(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}

// logBucket returns a string bucket label for a given positive number bucketed into
// powers of some arbitrary base. For base 8, for example, we would have buckets:
//
//	<1, [1-8), [8-64), [64-512), etc.
//
// Go only supports a few bases with math.Log*, so this func performs a change of base:
// log_b(x) = log_a(x) / log_a(b)
func logBucket(base int, num int64) string {
	if num < 1 {
		return "<1"
	}
	b := float64(base)
	logB := math.Log(float64(num)) / math.Log(b)
	bucketExpLo := math.Floor(logB)
	return fmt.Sprintf("[%d,%d)", int(math.Pow(b, bucketExpLo)), int(math.Pow(b, bucketExpLo+1)))
}
