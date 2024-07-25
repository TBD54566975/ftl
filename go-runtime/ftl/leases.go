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
	"github.com/alecthomas/types/optional"
)

// ErrLeaseHeld is returned when an attempt is made to acquire a lease that is
// already held.
var ErrLeaseHeld = fmt.Errorf("lease already held")

type leaseState struct {
	// mutex must be locked to access other fields.
	mutex *sync.Mutex

	// open is true if the lease has not been released and no error has occurred.
	open bool
	err  optional.Option[error]
}

type LeaseHandle struct {
	client modulecontext.LeaseClient
	module string
	key    []string
	state  *leaseState
}

// Err returns an error if the lease heartbeat fails.
func (l LeaseHandle) Err() error {
	l.state.mutex.Lock()
	defer l.state.mutex.Unlock()
	if err, ok := l.state.err.Get(); ok {
		return err //nolint:wrapcheck
	}
	return nil
}

// loggingKey mimics the full lease key for logging purposes.
// This helps us identify the lease in logs across runners and controllers.
func leaseKeyForLogs(module string, key []string) string {
	components := []string{"module", module}
	components = append(components, key...)
	return strings.Join(components, "/")
}

// Release attempts to release the lease.
//
// Will return an error if the heartbeat failed. In this situation there are no
// guarantees that the lease was held to completion.
func (l LeaseHandle) Release() error {
	l.state.mutex.Lock()
	defer l.state.mutex.Unlock()
	if !l.state.open {
		return l.Err()
	}
	err := l.client.Release(context.Background(), l.key)
	if err != nil {
		return err
	}
	l.state.open = false
	return nil
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
	logger.Tracef("Acquiring lease: %s", leaseKeyForLogs(module, key))
	err := client.Acquire(ctx, module, key, ttl)
	if err != nil {
		if errors.Is(err, ErrLeaseHeld) {
			return LeaseHandle{}, ErrLeaseHeld
		}
		logger.Warnf("Lease acquisition failed for %s: %s", leaseKeyForLogs(module, key), err)
		return LeaseHandle{}, err
	}
	lease := LeaseHandle{
		module: module,
		key:    key,
		client: client,
		state: &leaseState{
			open:  true,
			mutex: &sync.Mutex{},
		},
	}

	// Heartbeat the lease.
	go func() {
		for {
			logger.Tracef("Heartbeating lease: %s", leaseKeyForLogs(module, key))
			err := client.Heartbeat(ctx, module, key, ttl)
			if err == nil {
				time.Sleep(ttl / 2)
				continue
			}

			lease.state.mutex.Lock()
			defer lease.state.mutex.Unlock()
			if !lease.state.open {
				logger.Tracef("Lease heartbeat terminated for %s after being released", leaseKeyForLogs(module, key))
				return
			}
			// Notify the handle.
			logger.Warnf("Lease heartbeat terminated for %s: %s", leaseKeyForLogs(module, key), err)
			lease.state.open = false
			lease.state.err = optional.Some(err)
			return
		}
	}()
	return lease, nil
}

// newClient creates a new lease client
//
// It allows module context to override the client with a mock if appropriate
func newClient(ctx context.Context) modulecontext.LeaseClient {
	moduleCtx := modulecontext.FromContext(ctx).CurrentContext()
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

func (c *leaseClient) Heartbeat(_ context.Context, module string, key []string, ttl time.Duration) error {
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

func (c *leaseClient) Release(_ context.Context, _ []string) error {
	if err := c.stream.CloseRequest(); err != nil {
		return fmt.Errorf("close lease: %w", err)
	}
	if err := c.stream.CloseResponse(); err != nil {
		return fmt.Errorf("close lease: %w", err)
	}
	return nil
}
