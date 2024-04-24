package leases

import (
	"context"
	"errors"
	"time"
)

// ErrConflict is returned when a lease is already held.
var ErrConflict = errors.New("lease already held")

// Leaser is an interface for acquiring, renewing and releasing leases.
type Leaser interface {
	// AcquireLease attempts to acquire a new lease.
	//
	// A lease is a mechanism to ensure that only one controller is performing a
	// specific task at a time. The lease is held for a specific duration, and must
	// be renewed periodically to prevent it from expiring. If the lease expires, it
	// is released and another controller can acquire it.
	//
	// The lease is automatically renewed in the background, and can be
	// released by calling the [Lease.Release] method.
	//
	// This function will return [ErrConflict] if a lease is already held. The [ttl]
	// must be at _least_ 5 seconds.
	AcquireLease(ctx context.Context, key Key, ttl time.Duration) (Lease, error)
}

// Lease represents a lease that is held by a controller.
type Lease interface {
	Release() error
}
