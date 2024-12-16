package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	"github.com/jpillora/backoff"
	"github.com/titanous/json5"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

type benchCmd struct {
	Count       int            `short:"c" help:"Number of times to call the Verb in each thread." default:"128"`
	Parallelism int            `short:"j" help:"Number of concurrent benchmarks to create." default:"${numcpu}"`
	Wait        time.Duration  `short:"w" help:"Wait up to this elapsed time for the FTL cluster to become available." default:"1m"`
	Verb        reflection.Ref `arg:"" required:"" help:"Full path of Verb to call." predictor:"verbs"`
	Request     string         `arg:"" optional:"" help:"JSON5 request payload." default:"{}"`
}

func (c *benchCmd) Run(ctx context.Context, client ftlv1connect.VerbServiceClient) error {
	if err := rpc.Wait(ctx, backoff.Backoff{Max: time.Second * 2}, c.Wait, client); err != nil {
		return fmt.Errorf("FTL cluster did not become ready: %w", err)
	}
	logger := log.FromContext(ctx)
	request := map[string]any{}
	err := json5.Unmarshal([]byte(c.Request), &request)
	if err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	fmt.Printf("Starting benchmark\n")
	fmt.Printf("  Verb: %s\n", c.Verb)
	fmt.Printf("  Count: %d\n", c.Count)
	fmt.Printf("  Parallelism: %d\n", c.Parallelism)

	var errors int64
	var success int64
	wg := errgroup.Group{}
	timings := make([][]time.Duration, c.Parallelism)
	for job := range c.Parallelism {
		wg.Go(func() error {
			for range c.Count {
				start := time.Now()
				// otherwise, we have a match so call the verb
				_, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{
					Verb: c.Verb.ToProto(),
					Body: []byte(c.Request),
				}))
				if err != nil {
					// Only log error once.
					if atomic.AddInt64(&errors, 1) == 1 {
						logger.Errorf(err, "Error calling %s", c.Verb)
					}
				} else {
					atomic.AddInt64(&success, 1)
				}
				timings[job] = append(timings[job], time.Since(start))
			}
			return nil
		})
	}
	_ = wg.Wait() //nolint: errcheck

	// Display timing percentiles.
	var allTimings []time.Duration
	for _, t := range timings {
		allTimings = append(allTimings, t...)
	}
	sort.Slice(allTimings, func(i, j int) bool { return allTimings[i] < allTimings[j] })
	fmt.Printf("Results:\n")
	fmt.Printf("  Successes: %d\n", success)
	fmt.Printf("  Errors: %d\n", errors)
	fmt.Printf("Timing percentiles:\n")
	for p, t := range computePercentiles(allTimings) {
		fmt.Printf("  %d%%: %s\n", p, t)
	}
	fmt.Printf("Standard deviation: ±%v\n", computeStandardDeviation(allTimings))
	return nil
}

func computePercentiles(timings []time.Duration) map[int]time.Duration {
	percentiles := map[int]time.Duration{}
	for _, p := range []int{50, 90, 95, 99} {
		percentiles[p] = percentile(timings, p)
	}
	return percentiles
}

func percentile(timings []time.Duration, p int) time.Duration {
	if len(timings) == 0 {
		return 0
	}
	i := int(float64(len(timings)) * float64(p) / 100)
	return timings[i]
}

func computeStandardDeviation(timings []time.Duration) time.Duration {
	if len(timings) == 0 {
		return 0
	}

	var sum time.Duration
	for _, t := range timings {
		sum += t
	}
	mean := float64(sum) / float64(len(timings))

	var varianceSum float64
	for _, t := range timings {
		diff := float64(t) - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(timings))

	return time.Duration(math.Sqrt(variance))
}
