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
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

// ErrLeaseHeld is returned when an attempt is made to acquire a lease that is
// already held.
var ErrLeaseHeld = fmt.Errorf("lease already held")

type LeaseHandle struct {
	stream *connect.BidiStreamForClient[ftlv1.AcquireLeaseRequest, ftlv1.AcquireLeaseResponse]
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
	if err := l.stream.CloseRequest(); err != nil {
		return fmt.Errorf("close lease: %w", err)
	}
	if err := l.stream.CloseResponse(); err != nil {
		return fmt.Errorf("close lease: %w", err)
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
	client := rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
	stream := client.AcquireLease(ctx)

	module := Module()
	logger.Tracef("Acquiring lease")
	req := &ftlv1.AcquireLeaseRequest{Key: key, Module: module, Ttl: durationpb.New(ttl)}
	if err := stream.Send(req); err != nil {
		if connect.CodeOf(err) == connect.CodeResourceExhausted {
			return LeaseHandle{}, ErrLeaseHeld
		}
		logger.Warnf("Lease acquisition failed: %s", err)
		return LeaseHandle{}, fmt.Errorf("lease acquisition failed: %w", err)
	}
	// Wait for response.
	_, err := stream.Receive()
	if err != nil {
		if connect.CodeOf(err) == connect.CodeResourceExhausted {
			return LeaseHandle{}, ErrLeaseHeld
		}
		return LeaseHandle{}, fmt.Errorf("lease acquisition failed: %w", err)
	}

	lease := LeaseHandle{key: key, errMu: &sync.Mutex{}, stream: stream}
	// Heartbeat the lease.
	go func() {
		for {
			logger.Tracef("Heartbeating lease")
			req := &ftlv1.AcquireLeaseRequest{Key: key, Module: module, Ttl: durationpb.New(ttl)}
			err := stream.Send(req)
			if err == nil {
				time.Sleep(ttl / 2)
				continue
			}
			if errors.Is(err, io.EOF) {
				err = nil
			} else {
				logger.Warnf("Lease heartbeat terminated: %s", err)
			}
			// Notify the handle.
			lease.errMu.Lock()
			lease.err = err
			lease.errMu.Unlock()
			return
		}
	}()
	return lease, nil
}
