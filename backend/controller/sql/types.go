package sql

import (
	"database/sql/driver"
	"fmt"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
)

// Type is a database adapter type for schema.Type.
//
// It encodes to/from the protobuf representation of a Type.
type Type struct {
	schema.Type
}

func (t *Type) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		pb := &schemapb.Type{}
		if err := proto.Unmarshal(src, pb); err != nil {
			return err
		}
		t.Type = schema.TypeFromProto(pb)
		return nil
	default:
		return fmt.Errorf("cannot scan %T", src)
	}
}

func (t *Type) Value() (driver.Value, error) {
	data, err := proto.Marshal(t.Type.ToProto())
	if err != nil {
		return nil, err
	}
	return data, nil
}
