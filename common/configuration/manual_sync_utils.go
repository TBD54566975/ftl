//go:build !release

package configuration

import (
	"context"
	"net/url"
	"time"

	"github.com/alecthomas/atomic"

	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"
)

type manualSyncBlock struct {
	sync chan optional.Option[error]
}

// ManualSyncProvider prevents normal syncs by returning a very high sync interval
// when syncAndWait() is called, it starts returning a 0 sync interval  and then then blocks until sync completes.
// See why we didn't use mock clocks to schedule syncs here: https://github.com/TBD54566975/ftl/issues/2092
type ManualSyncProvider[R Role] struct {
	syncRequested atomic.Value[optional.Option[manualSyncBlock]]

	provider AsynchronousProvider[R]
}

var _ AsynchronousProvider[Secrets] = &ManualSyncProvider[Secrets]{}

func NewManualSyncProvider[R Role](provider AsynchronousProvider[R]) *ManualSyncProvider[R] {
	return &ManualSyncProvider[R]{
		provider: provider,
	}
}

func (a *ManualSyncProvider[R]) SyncAndWait() error {
	block := manualSyncBlock{
		sync: make(chan optional.Option[error]),
	}
	a.syncRequested.Store(optional.Some(block))
	err := <-block.sync
	if err, hasErr := err.Get(); hasErr {
		return err
	}
	return nil
}

func (a *ManualSyncProvider[R]) Role() R {
	return a.provider.Role()
}

func (a *ManualSyncProvider[R]) Key() string {
	return a.provider.Key()
}

func (a *ManualSyncProvider[R]) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	return a.provider.Store(ctx, ref, value)
}

func (a *ManualSyncProvider[R]) Delete(ctx context.Context, ref Ref) error {
	return a.provider.Delete(ctx, ref)
}

func (a *ManualSyncProvider[R]) SyncInterval() time.Duration {
	if _, ok := a.syncRequested.Load().Get(); ok {
		// sync now
		return 0
	}
	// prevent sync
	return time.Hour * 24 * 365
}

func (a *ManualSyncProvider[R]) Sync(ctx context.Context, entries []Entry, values *xsync.MapOf[Ref, SyncedValue]) error {
	err := a.provider.Sync(ctx, entries, values)

	if block, ok := a.syncRequested.Load().Get(); ok {
		a.syncRequested.Store(optional.None[manualSyncBlock]())
		if err == nil {
			block.sync <- optional.None[error]()
		} else {
			block.sync <- optional.Some(err)
		}
	}
	return err
}
