package sql

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
)

type NullRef = optional.Option[schema.Ref]
type NullUUID = optional.Option[uuid.UUID]
type NullLeaseKey = optional.Option[leases.Key]
type NullTime = optional.Option[time.Time]
type NullDuration = optional.Option[time.Duration]
type NullRunnerKey = optional.Option[model.RunnerKey]
type NullCronJobKey = optional.Option[model.CronJobKey]
type NullDeploymentKey = optional.Option[model.DeploymentKey]
type NullRequestKey = optional.Option[model.RequestKey]

var _ sql.Scanner = (*NullRunnerKey)(nil)
var _ driver.Valuer = (*NullRunnerKey)(nil)

var _ sql.Scanner = (*NullCronJobKey)(nil)
var _ driver.Valuer = (*NullCronJobKey)(nil)

var _ sql.Scanner = (*NullDeploymentKey)(nil)
var _ driver.Valuer = (*NullDeploymentKey)(nil)

var _ sql.Scanner = (*NullRequestKey)(nil)
var _ driver.Valuer = (*NullRequestKey)(nil)
