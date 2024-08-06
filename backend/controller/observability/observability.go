package observability

import (
	"errors"
	"fmt"
	"math"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	AsyncCalls *AsyncCallMetrics
	Deployment *DeploymentMetrics
	FSM        *FSMMetrics
	PubSub     *PubSubMetrics
)

func init() {
	var errs error
	var err error

	AsyncCalls, err = initAsyncCallMetrics()
	errs = errors.Join(errs, err)
	Deployment, err = initDeploymentMetrics()
	errs = errors.Join(errs, err)
	FSM, err = initFSMMetrics()
	errs = errors.Join(errs, err)
	PubSub, err = initPubSubMetrics()
	errs = errors.Join(errs, err)

	if err != nil {
		panic(fmt.Errorf("could not initialize controller metrics: %w", errs))
	}
}

//nolint:unparam
func handleInt64CounterError(counter string, err error, errs error) (metric.Int64Counter, error) {
	return noop.Int64Counter{}, errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
}

//nolint:unparam
func handleInt64UpDownCounterError(counter string, err error, errs error) (metric.Int64UpDownCounter, error) {
	return noop.Int64UpDownCounter{}, errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
}

func timeSinceMS(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}

// logBucket returns a string bucket label for a given positive number bucketed into
// powers of some arbitary base. For base 8, for example, we would have buckets:
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
	log_b := math.Log(float64(num)) / math.Log(b)
	bucketExpLo := math.Floor(log_b)
	return fmt.Sprintf("[%d,%d)", int(math.Pow(b, bucketExpLo)), int(math.Pow(b, bucketExpLo+1)))
}
