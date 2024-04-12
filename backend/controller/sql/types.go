package sql

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/model"
)

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
