package lease

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestLeaseService(t *testing.T) {
	t.Parallel()
	service := &service{
		leases: map[string]*time.Time{},
	}

	c1 := &leaseClient{
		leases:  map[string]*time.Time{},
		service: service,
	}
	c2 := &leaseClient{
		leases:  map[string]*time.Time{},
		service: service,
	}
	ok := c1.handleMessage([]string{"l1"}, time.Second)
	assert.True(t, ok)
	// Second client can't get the lease
	ok = c2.handleMessage([]string{"l1"}, time.Second)
	assert.False(t, ok)
	time.Sleep(time.Millisecond * 500)
	// First client can renew the lease
	ok = c1.handleMessage([]string{"l1"}, time.Second)
	assert.True(t, ok)
	// Second client can't get the lease
	ok = c2.handleMessage([]string{"l1"}, time.Second)
	assert.False(t, ok)
	time.Sleep(time.Second)
	time.Sleep(time.Millisecond)
	// lease has expired, client 2 can grab it now
	ok = c2.handleMessage([]string{"l1"}, time.Minute)
	assert.True(t, ok)
	// c1 should fail to renew the lease
	ok = c1.handleMessage([]string{"l1"}, time.Second)
	assert.False(t, ok)

}
