package leases

import (
	"context"
	"time"

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

func (f *FakeLeaser) AcquireLease(ctx context.Context, key Key, ttl time.Duration) (Lease, error) {
	if _, loaded := f.leases.LoadOrStore(key.String(), struct{}{}); loaded {
		return nil, ErrConflict
	}
	return &FakeLease{leaser: f, key: key}, nil
}

type FakeLease struct {
	leaser *FakeLeaser
	key    Key
}

func (f *FakeLease) Release() error {
	f.leaser.leases.Delete(f.key.String())
	return nil
}

func (f *FakeLease) String() string { return f.key.String() }
