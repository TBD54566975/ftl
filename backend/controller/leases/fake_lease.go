package leases

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"
)

func NewFakeLeaser() *FakeLeaser {
	return &FakeLeaser{
		leases: xsync.NewMapOf[string, struct{}](),
	}
}

var _ Leaser = (*FakeLeaser)(nil)

// FakeLeaser is a fake implementation of the [Leaser] interface.
type FakeLeaser struct {
	leases *xsync.MapOf[string, struct{}]
}

func (f *FakeLeaser) AcquireLease(ctx context.Context, key Key, ttl time.Duration, metadata optional.Option[any]) (Lease, context.Context, error) {
	if _, loaded := f.leases.LoadOrStore(key.String(), struct{}{}); loaded {
		return nil, nil, ErrConflict
	}
	leaseCtx, cancelCtx := context.WithCancel(ctx)
	return &FakeLease{
		leaser:    f,
		key:       key,
		cancelCtx: cancelCtx,
	}, leaseCtx, nil
}

func (f *FakeLeaser) GetLeaseInfo(ctx context.Context, key Key, metadata any) (expiry time.Time, err error) {
	if _, ok := f.leases.Load(key.String()); ok {
		return time.Now().Add(time.Minute), nil
	}
	return time.Time{}, fmt.Errorf("not found")
}

type FakeLease struct {
	leaser    *FakeLeaser
	key       Key
	cancelCtx context.CancelFunc
}

func (f *FakeLease) Release() error {
	f.leaser.leases.Delete(f.key.String())
	f.cancelCtx()
	return nil
}

func (f *FakeLease) String() string { return f.key.String() }
