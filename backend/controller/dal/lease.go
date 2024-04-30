package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/internal/log"
)

const leaseRenewalInterval = time.Second * 2

var _ leases.Leaser = (*DAL)(nil)

// Lease represents a lease that is held by a controller.
type Lease struct {
	key            leases.Key
	idempotencyKey uuid.UUID
	db             *sql.DB
	ttl            time.Duration
	errch          chan error
	release        chan bool
	leak           bool // For testing.
}

func (l *Lease) String() string {
	return fmt.Sprintf("%s:%s", l.key, l.idempotencyKey)
}

// Periodically renew the lease until it is released.
func (l *Lease) renew(ctx context.Context) {
	defer close(l.errch)
	logger := log.FromContext(ctx).Scope("lease(" + l.key.String() + ")")
	logger.Debugf("Acquired lease %s", l.key)
	for {
		select {
		case <-time.After(leaseRenewalInterval):
			logger.Tracef("Renewing lease %s", l.key)
			ctx, cancel := context.WithTimeout(ctx, leaseRenewalInterval)
			_, err := l.db.RenewLease(ctx, l.ttl, l.idempotencyKey, l.key)
			cancel()

			if err != nil {
				logger.Errorf(err, "Failed to renew lease %s", l.key)
				l.errch <- translatePGError(err)
				return
			}

		case <-l.release:
			if l.leak { // For testing.
				return
			}
			logger.Debugf("Releasing lease %s", l.key)
			_, err := l.db.ReleaseLease(ctx, l.idempotencyKey, l.key)
			l.errch <- translatePGError(err)
			return
		}
	}
}

func (l *Lease) Release() error {
	close(l.release)
	return <-l.errch
}

func (d *DAL) AcquireLease(ctx context.Context, key leases.Key, ttl time.Duration) (leases.Lease, error) {
	if ttl < time.Second*5 {
		return nil, fmt.Errorf("lease TTL must be at least 5 seconds")
	}
	idempotencyKey, err := d.db.NewLease(ctx, key, ttl)
	if err != nil {
		return nil, translatePGError(err)
	}
	return d.newLease(ctx, key, idempotencyKey, ttl), nil
}

func (d *DAL) newLease(ctx context.Context, key leases.Key, idempotencyKey uuid.UUID, ttl time.Duration) *Lease {
	lease := &Lease{
		idempotencyKey: idempotencyKey,
		key:            key,
		db:             d.db,
		ttl:            ttl,
		release:        make(chan bool),
		errch:          make(chan error, 1),
	}
	go lease.renew(ctx)
	return lease
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
