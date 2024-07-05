package leases

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
)

func TestLease(t *testing.T) {
	ctx := ftltest.Context(t)
	// test that we can acquire a lease in a test environment
	wg := errgroup.Group{}
	wg.Go(func() error {
		return Acquire(ctx)
	})

	// test that we get an error acquiring another lease at the same time
	time.Sleep(1 * time.Second)
	err := Acquire(ctx)
	assert.Error(t, err, "expected error for acquiring lease while another is held")

	assert.NoError(t, wg.Wait(), "expected no error acquiring the initial lease")
}
