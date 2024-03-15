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
type NullDeploymentName = optional.Option[model.DeploymentName]

var _ sql.Scanner = (*NullRunnerKey)(nil)
var _ driver.Valuer = (*NullRunnerKey)(nil)

var _ sql.Scanner = (*NullDeploymentName)(nil)
var _ driver.Valuer = (*NullDeploymentName)(nil)
