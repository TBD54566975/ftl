package observability

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/alecthomas/types/optional"
)

var (
	AsyncCalls *AsyncCallMetrics
	Calls      *CallMetrics
	Deployment *DeploymentMetrics
	FSM        *FSMMetrics
	Ingress    *IngressMetrics
	PubSub     *PubSubMetrics
	Cron       *CronMetrics
	Controller *ControllerTracing
	Timeline   *TimelineMetrics
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
	Ingress, err = initIngressMetrics()
	errs = errors.Join(errs, err)
	PubSub, err = initPubSubMetrics()
	errs = errors.Join(errs, err)
	Cron, err = initCronMetrics()
	errs = errors.Join(errs, err)
	Controller = initControllerTracing()
	Timeline, err = initTimelineMetrics()
	errs = errors.Join(errs, err)

	if errs != nil {
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
// The buckets are then demarcated by `min` and `max`, such that any `num` < `base`^`min`
// will be bucketed together into the min bucket, and similarly, any `num` >= `base`^`max`
// will go in the `max` bucket. This constrains output cardinality by chopping the long
// tails on both ends of the normal distribution and lumping them all into terminal
// buckets. When `min` and `max` are not provided, the effective `min` is 0, and there is
// no max.
//
// Go only supports a few bases with math.Log*, so this func performs a change of base:
// log_b(x) = log_a(x) / log_a(b)
func logBucket(base int, num int64, optMin, optMax optional.Option[int]) string {
	if num < 1 {
		return "<1"
	}
	b := float64(base)

	// Check max
	maxBucket, ok := optMax.Get()
	if ok {
		maxThreshold := int64(math.Pow(b, float64(maxBucket)))
		if num >= maxThreshold {
			return fmt.Sprintf(">=%d", maxThreshold)
		}
	}

	// Check min
	minBucket, ok := optMin.Get()
	if ok {
		minThreshold := int64(math.Pow(b, float64(minBucket)))
		if num < minThreshold {
			return fmt.Sprintf("<%d", minThreshold)
		}
	}

	logB := math.Log(float64(num)) / math.Log(b)
	bucketExpLo := math.Floor(logB)

	return fmt.Sprintf("[%d,%d)", int(math.Pow(b, bucketExpLo)), int(math.Pow(b, bucketExpLo+1)))
}
