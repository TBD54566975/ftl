package scheduledtask

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/benbjohnson/clock"
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
)

func TestCron(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{Level: log.Debug}))
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	var singletonCount atomic.Int64
	var multiCount atomic.Int64

	type controller struct {
		controller dal.Controller
		cron       *Scheduler
	}

	controllers := []*controller{
		{controller: dal.Controller{Key: model.NewControllerKey()}},
		{controller: dal.Controller{Key: model.NewControllerKey()}},
		{controller: dal.Controller{Key: model.NewControllerKey()}},
		{controller: dal.Controller{Key: model.NewControllerKey()}},
	}

	clock := clock.NewMock()

	for _, c := range controllers {
		c := c
		c.cron = NewForTesting(ctx, c.controller.Key, DALFunc(func(ctx context.Context, all bool) ([]dal.Controller, error) {
			return slices.Map(controllers, func(c *controller) dal.Controller { return c.controller }), nil
		}), clock)
		c.cron.Singleton(backoff.Backoff{}, func(ctx context.Context) (time.Duration, error) {
			singletonCount.Add(1)
			return time.Second, nil
		})
		c.cron.Parallel(backoff.Backoff{}, func(ctx context.Context) (time.Duration, error) {
			multiCount.Add(1)
			return time.Second, nil
		})
	}

	clock.Add(time.Second * 6)

	assert.True(t, singletonCount.Load() >= 5 && singletonCount.Load() < 10, "expected singletonCount to be >= 5 but was %d", singletonCount.Load())
	assert.True(t, multiCount.Load() >= 20 && multiCount.Load() < 30, "expected multiCount to be >= 20 but was %d", multiCount.Load())
}
