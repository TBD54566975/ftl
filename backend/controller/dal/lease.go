package dal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/internal/log"
)

var _ leases.Leaser = (*DAL)(nil)

// Lease represents a lease that is held by a controller.
type Lease struct {
	key            leases.Key
	idempotencyKey uuid.UUID
	db             sql.DBI
	ttl            time.Duration
	errch          chan error
	release        chan bool
	cancelCtx      context.CancelFunc // cancels context created for lease owner
	leak           bool               // For testing.
}

func (l *Lease) String() string {
	return fmt.Sprintf("%s:%s", l.key, l.idempotencyKey)
}

// Periodically renew the lease until it is released.
func (l *Lease) renew(ctx context.Context) {
	defer close(l.errch)
	leaseRenewalInterval := l.ttl / 2
	logger := log.FromContext(ctx).Scope("lease:" + l.key.String())
	logger.Debugf("Acquired lease")
	for {
		select {
		case <-time.After(leaseRenewalInterval):
			logger.Tracef("Renewing lease")
			ctx, cancel := context.WithTimeout(ctx, leaseRenewalInterval)
			_, err := l.db.RenewLease(ctx, l.ttl, l.idempotencyKey, l.key)
			cancel()

			if err != nil {
				err = translatePGError(err)
				if errors.Is(err, ErrNotFound) {
					logger.Warnf("Lease expired")
				} else {
					logger.Errorf(err, "Failed to renew lease %s", l.key)
				}
				l.errch <- err
				l.cancelCtx()
				return
			}

		case <-l.release:
			if l.leak { // For testing.
				return
			}
			logger.Debugf("Releasing lease")
			_, err := l.db.ReleaseLease(ctx, l.idempotencyKey, l.key)
			l.errch <- translatePGError(err)
			l.cancelCtx()
			return
		}
	}
}

func (l *Lease) Release() error {
	close(l.release)
	return <-l.errch
}

// AcquireLease acquires a lease for the given key.
//
// Will return ErrConflict if the lease is already held by another controller.
func (d *DAL) AcquireLease(ctx context.Context, key leases.Key, ttl time.Duration, metadata optional.Option[any]) (leases.Lease, context.Context, error) {
	if ttl < time.Second*5 {
		return nil, nil, fmt.Errorf("lease TTL must be at least 5 seconds")
	}
	var metadataBytes []byte
	if md, ok := metadata.Get(); ok {
		var err error
		metadataBytes, err = json.Marshal(md)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal lease metadata: %w", err)
		}
	}
	idempotencyKey, err := d.db.NewLease(ctx, key, ttl, metadataBytes)
	if err != nil {
		return nil, nil, translatePGError(err)
	}
	leaseCtx, lease := d.newLease(ctx, key, idempotencyKey, ttl)
	return leaseCtx, lease, nil
}

func (d *DAL) newLease(ctx context.Context, key leases.Key, idempotencyKey uuid.UUID, ttl time.Duration) (*Lease, context.Context) {
	leaseCtx, cancelCtx := context.WithCancel(ctx)
	lease := &Lease{
		idempotencyKey: idempotencyKey,
		key:            key,
		db:             d.db,
		ttl:            ttl,
		release:        make(chan bool),
		errch:          make(chan error, 1),
		cancelCtx:      cancelCtx,
	}
	go lease.renew(ctx)
	return lease, leaseCtx
}

// GetLeaseInfo returns the metadata and expiry time for the lease with the given key.
//
// metadata should be a pointer to the type that metadata should be unmarshaled into.
func (d *DAL) GetLeaseInfo(ctx context.Context, key leases.Key, metadata any) (expiry time.Time, err error) {
	l, err := d.db.GetLeaseInfo(ctx, key)
	if err != nil {
		return expiry, translatePGError(err)
	}
	if err := json.Unmarshal(l.Metadata, metadata); err != nil {
		return expiry, fmt.Errorf("could not unmarshal lease metadata: %w", err)
	}
	return l.ExpiresAt, nil
}

// ExpireLeases expires (deletes) all leases that have expired.
func (d *DAL) ExpireLeases(ctx context.Context) error {
	count, err := d.db.ExpireLeases(ctx)
	// TODO: Return and log the actual lease keys?
	if count > 0 {
		log.FromContext(ctx).Warnf("Expired %d leases", count)
	}
	return translatePGError(err)
}
