package ftl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/durationpb"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/modulecontext"
	"github.com/TBD54566975/ftl/internal/rpc"
)

// ErrLeaseHeld is returned when an attempt is made to acquire a lease that is
// already held.
var ErrLeaseHeld = fmt.Errorf("lease already held")

type LeaseHandle struct {
	client modulecontext.LeaseClient
	key    []string
	errMu  *sync.Mutex
	err    error
}

// Err returns an error if the lease heartbeat fails.
func (l LeaseHandle) Err() error {
	l.errMu.Lock()
	defer l.errMu.Unlock()
	return l.err
}

// Release attempts to release the lease.
//
// Will return an error if the heartbeat failed. In this situation there are no
// guarantees that the lease was held to completion.
func (l LeaseHandle) Release() error {
	l.errMu.Lock()
	defer l.errMu.Unlock()
	err := l.client.Release(context.Background(), l.key)
	if err != nil {
		return err
	}
	return l.err
}

// Lease acquires a new exclusive [lease] on a resource uniquely identified by [key].
//
// The [ttl] defines the time after which the lease will be released if no
// heartbeat has been received. It must be >= 5s.
//
// Each [key] is scoped to the module that acquires the lease.
//
// Returns [ErrLeaseHeld] if the lease is already held.
//
// [lease]: https://hackmd.io/@ftl/Sym_GKEb0
func Lease(ctx context.Context, ttl time.Duration, key ...string) (LeaseHandle, error) {
	logger := log.FromContext(ctx).Scope("lease(" + strings.Join(key, "/"))
	client := newClient(ctx)

	module := reflection.Module()
	logger.Tracef("Acquiring lease")
	err := client.Acquire(ctx, module, key, ttl)
	if err != nil {
		if err != ErrLeaseHeld {
			return LeaseHandle{}, ErrLeaseHeld
		}
		logger.Warnf("Lease acquisition failed: %s", err)
		return LeaseHandle{}, err
	}

	lease := LeaseHandle{key: key, errMu: &sync.Mutex{}, client: client}
	// Heartbeat the lease.
	go func() {
		for {
			logger.Tracef("Heartbeating lease")
			err := client.Heartbeat(ctx, module, key, ttl)
			if err == nil {
				time.Sleep(ttl / 2)
				continue
			}
			logger.Warnf("Lease heartbeat terminated: %s", err)

			// Notify the handle.
			lease.errMu.Lock()
			lease.err = err
			lease.errMu.Unlock()
			return
		}
	}()
	return lease, nil
}

// newClient creates a new lease client
//
// It allows module context to override the client with a mock if appropriate
func newClient(ctx context.Context) modulecontext.LeaseClient {
	moduleCtx := modulecontext.FromContext(ctx)
	if mock, ok := moduleCtx.MockLeaseClient().Get(); ok {
		return mock
	}
	return &leaseClient{}
}

// leaseClient is a client that coordinates leases with the controller
//
// This is used in all non-unit tests environements
type leaseClient struct {
	stream *connect.BidiStreamForClient[ftlv1.AcquireLeaseRequest, ftlv1.AcquireLeaseResponse]
}

var _ modulecontext.LeaseClient = &leaseClient{}

func (c *leaseClient) Acquire(ctx context.Context, module string, key []string, ttl time.Duration) error {
	c.stream = rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx).AcquireLease(ctx)
	req := &ftlv1.AcquireLeaseRequest{Key: key, Module: module, Ttl: durationpb.New(ttl)}
	if err := c.stream.Send(req); err != nil {
		if connect.CodeOf(err) == connect.CodeResourceExhausted {
			return ErrLeaseHeld
		}
		return fmt.Errorf("lease acquisition failed: %w", err)
	}
	// Wait for response.
	_, err := c.stream.Receive()
	if err == nil {
		return nil
	}
	if connect.CodeOf(err) == connect.CodeResourceExhausted {
		return ErrLeaseHeld
	}
	return fmt.Errorf("lease acquisition failed: %w", err)
}

func (c *leaseClient) Heartbeat(ctx context.Context, module string, key []string, ttl time.Duration) error {
	req := &ftlv1.AcquireLeaseRequest{Key: key, Module: module, Ttl: durationpb.New(ttl)}
	err := c.stream.Send(req)
	if err == nil {
		return nil
	}
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func (c *leaseClient) Release(ctx context.Context, key []string) error {
	if err := c.stream.CloseRequest(); err != nil {
		return fmt.Errorf("close lease: %w", err)
	}
	if err := c.stream.CloseResponse(); err != nil {
		return fmt.Errorf("close lease: %w", err)
	}
	return nil
}
