package leases

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"
)

func NewFakeLeaser() *FakeLeaser {
	return &FakeLeaser{
		leases: xsync.NewMapOf[string, *FakeLease](),
	}
}

var _ Leaser = (*FakeLeaser)(nil)

// FakeLeaser is a fake implementation of the [Leaser] interface.
type FakeLeaser struct {
	leases *xsync.MapOf[string, *FakeLease]
}

func (f *FakeLeaser) AcquireLease(ctx context.Context, key Key, ttl time.Duration, metadata optional.Option[any]) (Lease, context.Context, error) {
	leaseCtx, cancelCtx := context.WithCancel(ctx)
	newLease := &FakeLease{
		leaser:    f,
		key:       key,
		metadata:  metadata,
		cancelCtx: cancelCtx,
		ttl:       ttl,
	}
	if _, loaded := f.leases.LoadOrStore(key.String(), newLease); loaded {
		cancelCtx()
		return nil, nil, ErrConflict
	}
	return newLease, leaseCtx, nil
}

func (f *FakeLeaser) GetLeaseInfo(ctx context.Context, key Key, metadata any) (expiry time.Time, err error) {
	if lease, ok := f.leases.Load(key.String()); ok {
		if md, ok := lease.metadata.Get(); ok && metadata != nil {
			// set metadata value
			metaValue := reflect.ValueOf(metadata)
			if metaValue.Kind() != reflect.Ptr || metaValue.IsNil() {
				return time.Time{}, fmt.Errorf("metadata must be a non-nil pointer")
			}
			if !metaValue.Elem().CanSet() {
				return time.Time{}, fmt.Errorf("cannot set metadata value")
			}
			mdValue := reflect.ValueOf(md)
			if mdValue.Type() != metaValue.Elem().Type() {
				return time.Time{}, fmt.Errorf("type mismatch between metadata and md")
			}
			metaValue.Elem().Set(mdValue)
		}
		return time.Now().Add(lease.ttl), nil
	}
	return time.Time{}, fmt.Errorf("not found")
}

type FakeLease struct {
	leaser    *FakeLeaser
	key       Key
	cancelCtx context.CancelFunc
	metadata  optional.Option[any]
	ttl       time.Duration
}

func (f *FakeLease) Release() error {
	f.leaser.leases.Delete(f.key.String())
	f.cancelCtx()
	return nil
}

func (f *FakeLease) String() string { return f.key.String() }
