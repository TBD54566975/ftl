package leases

import (
	"context"
	"errors"
	"fmt"
	"time"

	leasepb "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
	"github.com/block/ftl/internal/rpc"
)

var _ Leaser = (*clientLeaser)(nil)
var _ Lease = (*clientLease)(nil)

func NewClientLeaser(ctx context.Context) Leaser {
	return &clientLeaser{
		client: rpc.ClientFromContext[leasepbconnect.LeaseServiceClient](ctx),
	}
}

type clientLeaser struct {
	client leasepbconnect.LeaseServiceClient
}

func (c clientLeaser) AcquireLease(ctx context.Context, key Key, ttl time.Duration) (Lease, context.Context, error) {
	if ttl.Seconds() < 5 {
		return nil, nil, errors.New("ttl must be at least 5 seconds")
	}
	lease := c.client.AcquireLease(ctx)
	// Send the initial request to acquire the lease.
	err := lease.Send(&leasepb.AcquireLeaseRequest{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send acquire lease request: %w", err)
	}
	_, err = lease.Receive()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send receive lease response: %w", err)
	}
	// We have got the lease, we need a goroutine to keep renewing the lease.
	ret := &clientLease{}
	ctx, cancel := context.WithCancel(ctx)
	done := func() {
		cancel()
		_ = lease.CloseResponse() //nolint:errcheck
		_ = lease.CloseRequest()  //nolint:errcheck
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				done()
				return
			case <-time.After(ttl / 2):
				err := lease.Send(&leasepb.AcquireLeaseRequest{})
				if err != nil {
					done()
					return
				}
				_, err = lease.Receive()
				if err != nil {
					done()
					return
				}
			}

		}
	}()
	return ret, ctx, nil
}

type clientLease struct {
	done func()
}

func (c *clientLease) Release() error {
	c.done()
	return nil
}
