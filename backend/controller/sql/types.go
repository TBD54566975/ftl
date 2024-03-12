package sql

import (
	"time"

	"github.com/alecthomas/types/optional"
)

type NullTime = optional.Option[time.Time]
type NullDuration = optional.Option[time.Duration]
