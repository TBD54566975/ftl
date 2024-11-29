package sqltypes

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/schema"
)

type Duration time.Duration

func (d Duration) Value() (driver.Value, error) {
	return time.Duration(d).String(), nil
}

func (d *Duration) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		// Convert format of hh:mm:ss into format parseable by time.ParseDuration()
		v = strings.Replace(v, ":", "h", 1)
		v = strings.Replace(v, ":", "m", 1)
		v += "s"
		dur, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("failed to parse duration %q: %w", v, err)
		}
		*d = Duration(dur)
		return nil
	default:
		return fmt.Errorf("cannot scan duration %v", value)
	}
}

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
			return fmt.Errorf("could not unmarshal type: %w", err)
		}
		t.Type = schema.TypeFromProto(pb)
		return nil
	default:
		return fmt.Errorf("cannot scan %T", src)
	}
}

func (t Type) Value() (driver.Value, error) {
	data, err := proto.Marshal(schema.TypeToProto(t.Type))
	if err != nil {
		return nil, fmt.Errorf("could not marshal type: %w", err)
	}
	return data, nil
}

type OptionalTime = optional.Option[time.Time]
