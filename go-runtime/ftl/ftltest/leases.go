package ftltest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/internal/modulecontext"
)

// leaseClient is a simple in-memory lease client for testing.
//
// It does not include any checks on module names, as it assumes that all leases are within the module being tested
type leaseClient struct {
	lock      sync.Mutex
	deadlines map[string]time.Time
}

var _ modulecontext.LeaseClient = &leaseClient{}

func newLeaseClient() *leaseClient {
	return &leaseClient{
		deadlines: make(map[string]time.Time),
	}
}

func keyForKeys(keys []string) string {
	return strings.Join(keys, "\n")
}

func (c *leaseClient) Acquire(ctx context.Context, module string, key []string, ttl time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	k := keyForKeys(key)
	if deadline, ok := c.deadlines[k]; ok {
		if time.Now().Before(deadline) {
			return ftl.ErrLeaseHeld
		}
	}

	c.deadlines[k] = time.Now().Add(ttl)
	return nil
}

func (c *leaseClient) Heartbeat(ctx context.Context, module string, key []string, ttl time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	k := keyForKeys(key)
	if deadline, ok := c.deadlines[k]; ok {
		if !time.Now().Before(deadline) {
			return fmt.Errorf("could not heartbeat expired lease")
		}
		c.deadlines[k] = time.Now().Add(ttl)
		return nil
	}
	return fmt.Errorf("could not heartbeat lease: no active lease found")
}

func (c *leaseClient) Release(ctx context.Context, key []string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	k := keyForKeys(key)
	if deadline, ok := c.deadlines[k]; ok {
		if !time.Now().Before(deadline) {
			return fmt.Errorf("could not release lease after timeout")
		}
		delete(c.deadlines, k)
		return nil
	}
	return fmt.Errorf("could not release lease: no active lease found")
}
