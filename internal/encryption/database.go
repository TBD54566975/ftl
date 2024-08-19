package encryption

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/alecthomas/types/optional"
)

var _ Encrypted = &EncryptedColumn[TimelineSubKey]{}

// EncryptedColumn is a type that represents an encrypted column.
//
// It can be used by sqlc to map to/from a bytea column in the database.
type EncryptedColumn[SK SubKey] []byte

var _ driver.Valuer = &EncryptedColumn[TimelineSubKey]{}
var _ sql.Scanner = &EncryptedColumn[TimelineSubKey]{}

func (e *EncryptedColumn[SK]) SubKey() string { var sk SK; return sk.SubKey() }
func (e *EncryptedColumn[SK]) Bytes() []byte  { return *e }
func (e *EncryptedColumn[SK]) Set(b []byte)   { *e = b }
func (e *EncryptedColumn[SK]) Value() (driver.Value, error) {
	return []byte(*e), nil
}

func (e *EncryptedColumn[SK]) Scan(src interface{}) error {
	if src == nil {
		*e = nil
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", src)
	}
	*e = b
	return nil
}

type EncryptedTimelineColumn = EncryptedColumn[TimelineSubKey]
type EncryptedAsyncColumn = EncryptedColumn[AsyncSubKey]

type OptionalEncryptedTimelineColumn = optional.Option[EncryptedTimelineColumn]
type OptionalEncryptedAsyncColumn = optional.Option[EncryptedAsyncColumn]

// SubKey is an interface for types that specify their own encryption subkey.
type SubKey interface{ SubKey() string }

// TimelineSubKey is a type that represents the subkey for logs.
type TimelineSubKey struct{}

func (TimelineSubKey) SubKey() string { return "logs" }

// AsyncSubKey is a type that represents the subkey for async.
type AsyncSubKey struct{}

func (AsyncSubKey) SubKey() string { return "async" }
