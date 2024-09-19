package api

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
type EncryptedColumn[SK SubKey] struct{ data []byte }

var _ driver.Valuer = &EncryptedColumn[TimelineSubKey]{}
var _ sql.Scanner = &EncryptedColumn[TimelineSubKey]{}

func (e *EncryptedColumn[SK]) SubKey() string              { var sk SK; return sk.SubKey() }
func (e *EncryptedColumn[SK]) Bytes() []byte               { return e.data }
func (e *EncryptedColumn[SK]) Set(b []byte)                { e.data = b }
func (e EncryptedColumn[SK]) Value() (driver.Value, error) { return e.data, nil }
func (e *EncryptedColumn[SK]) GoString() string {
	return fmt.Sprintf("EncryptedColumn[%s](%d bytes)", e.SubKey(), len(e.data))
}

func (e *EncryptedColumn[SK]) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", src)
	}
	e.data = b
	return nil
}

type EncryptedTimelineColumn = EncryptedColumn[TimelineSubKey]
type EncryptedAsyncColumn = EncryptedColumn[AsyncSubKey]
type EncryptedIdentityKey = EncryptedColumn[IdentityKeySubKey]

type OptionalEncryptedTimelineColumn = optional.Option[EncryptedTimelineColumn]
type OptionalEncryptedAsyncColumn = optional.Option[EncryptedAsyncColumn]

// SubKey is an interface for types that specify their own encryption subkey.
type SubKey interface{ SubKey() string }

// TimelineSubKey is a type that represents the subkey for logs.
type TimelineSubKey struct{}

func (TimelineSubKey) SubKey() string { return "timeline" }

// AsyncSubKey is a type that represents the subkey for async.
type AsyncSubKey struct{}

func (AsyncSubKey) SubKey() string { return "async" }

// IdentityKeySubKey is a type that represents the subkey for identity keys.
type IdentityKeySubKey struct{}

func (IdentityKeySubKey) SubKey() string { return "identity" }
