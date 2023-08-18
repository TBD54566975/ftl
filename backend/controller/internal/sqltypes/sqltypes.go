package sqltypes

import (
	"database/sql/driver"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

type NullKey = types.Option[Key]

// FromOption converts a types.Option[~ulid.ULID] to a NullKey.
func FromOption[T ~[16]byte](o types.Option[T]) NullKey {
	if v, ok := o.Get(); ok {
		return SomeKey(Key(v))
	}
	return NoneKey()
}
func SomeKey(key Key) NullKey { return types.Some(key) }
func NoneKey() NullKey        { return types.None[Key]() }

// Key is a ULID that can be used as a column in a database.
type Key ulid.ULID

func (u Key) Value() (driver.Value, error) {
	bytes := u[:]
	return bytes, nil
}

func (u *Key) Scan(src interface{}) error {
	id, err := uuid.Parse(src.(string))
	if err != nil {
		return errors.WithStack(err)
	}
	*u = Key(id)
	return nil
}

func (u *Key) UnmarshalText(text []byte) error {
	id, err := uuid.ParseBytes(text)
	if err != nil {
		return errors.WithStack(err)
	}
	*u = Key(id)
	return nil
}

type NullTime = types.Option[time.Time]
type NullDuration = types.Option[time.Duration]
