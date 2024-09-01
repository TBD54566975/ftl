//go:build !release

package providerstest

import (
	"context"
	"net/url"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/internal/configuration"
)

type manualSyncBlock struct {
	sync chan optional.Option[error]
}

// ManualSyncProvider prevents normal syncs by returning a very high sync interval
// when syncAndWait() is called, it starts returning a 0 sync interval  and then then blocks until sync completes.
// See why we didn't use mock clocks to schedule syncs here: https://github.com/TBD54566975/ftl/issues/2092
type ManualSyncProvider[R configuration.Role] struct {
	syncRequested atomic.Value[optional.Option[manualSyncBlock]]

	provider configuration.AsynchronousProvider[R]
}

var _ configuration.AsynchronousProvider[configuration.Secrets] = &ManualSyncProvider[configuration.Secrets]{}

func NewManualSyncProvider[R configuration.Role](provider configuration.AsynchronousProvider[R]) *ManualSyncProvider[R] {
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
		return err //nolint:wrapcheck
	}
	return nil
}

func (a *ManualSyncProvider[R]) Role() R {
	return a.provider.Role()
}

func (a *ManualSyncProvider[R]) Key() configuration.ProviderKey {
	return a.provider.Key()
}

func (a *ManualSyncProvider[R]) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	return a.provider.Store(ctx, ref, value) //nolint:wrapcheck
}

func (a *ManualSyncProvider[R]) Delete(ctx context.Context, ref configuration.Ref) error {
	return a.provider.Delete(ctx, ref) //nolint:wrapcheck
}

func (a *ManualSyncProvider[R]) SyncInterval() time.Duration {
	if _, ok := a.syncRequested.Load().Get(); ok {
		// sync now
		return 0
	}
	// prevent sync
	return time.Hour * 24 * 365
}

func (a *ManualSyncProvider[R]) Sync(ctx context.Context, entries []configuration.Entry, values *xsync.MapOf[configuration.Ref, configuration.SyncedValue]) error {
	err := a.provider.Sync(ctx, entries, values)

	if block, ok := a.syncRequested.Load().Get(); ok {
		a.syncRequested.Store(optional.None[manualSyncBlock]())
		if err == nil {
			block.sync <- optional.None[error]()
		} else {
			block.sync <- optional.Some(err)
		}
	}
	return err //nolint:wrapcheck
}
