package sql

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
)

type NullType = optional.Option[Type]
type NullRef = optional.Option[schema.RefKey]
type NullUUID = optional.Option[uuid.UUID]
type NullLeaseKey = optional.Option[leases.Key]
type NullTime = optional.Option[time.Time]
type NullDuration = optional.Option[time.Duration]
type NullRunnerKey = optional.Option[model.RunnerKey]
type NullCronJobKey = optional.Option[model.CronJobKey]
type NullDeploymentKey = optional.Option[model.DeploymentKey]
type NullRequestKey = optional.Option[model.RequestKey]
type NullTopicKey = optional.Option[model.TopicKey]
type NullSubscriptionKey = optional.Option[model.SubscriptionKey]
type NullSubscriberKey = optional.Option[model.SubscriberKey]
type NullTopicEventKey = optional.Option[model.TopicEventKey]

var _ sql.Scanner = (*NullRunnerKey)(nil)
var _ driver.Valuer = (*NullRunnerKey)(nil)

var _ sql.Scanner = (*NullCronJobKey)(nil)
var _ driver.Valuer = (*NullCronJobKey)(nil)

var _ sql.Scanner = (*NullDeploymentKey)(nil)
var _ driver.Valuer = (*NullDeploymentKey)(nil)

var _ sql.Scanner = (*NullRequestKey)(nil)
var _ driver.Valuer = (*NullRequestKey)(nil)

var _ sql.Scanner = (*NullTopicKey)(nil)
var _ driver.Valuer = (*NullTopicKey)(nil)

var _ sql.Scanner = (*NullSubscriptionKey)(nil)
var _ driver.Valuer = (*NullSubscriptionKey)(nil)

var _ sql.Scanner = (*NullSubscriberKey)(nil)
var _ driver.Valuer = (*NullSubscriberKey)(nil)

var _ sql.Scanner = (*NullTopicEventKey)(nil)
var _ driver.Valuer = (*NullTopicEventKey)(nil)

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
