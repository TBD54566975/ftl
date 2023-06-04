package sqltypes

import (
	"database/sql/driver"

	"github.com/alecthomas/errors"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// Key is a ULID that can be used as a column in a database.
type Key ulid.ULID

func (u Key) Value() (driver.Value, error) {
	bytes := u[:]
	return bytes, nil
}

var _ driver.Valuer = (*Key)(nil)

func (u *Key) Scan(src interface{}) error {
	id, err := uuid.Parse(src.(string))
	if err != nil {
		return errors.WithStack(err)
	}
	*u = Key(id)
	return nil
}

func (u Key) ULID() ulid.ULID {
	return ulid.ULID(u)
}
