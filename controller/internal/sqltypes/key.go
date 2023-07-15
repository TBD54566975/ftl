package sqltypes

import (
	"database/sql/driver"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

type NullKey = types.Option[Key]

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
