package controller

import (
	"net/url"
	"time"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/internal/eventstream"
	"github.com/TBD54566975/ftl/internal/model"
)

var _ eventstream.View[Runner] = (*RunnerState)(nil)

type RunnerState struct {
	runners map[string]Runner
}

func (r *RunnerState) Entry(s string) optional.Option[Runner] {
	result, ok := r.runners[s]
	if ok {
		return optional.Some(result)
	}
	return optional.None[Runner]()
}

func (r *RunnerState) Entries() []Runner {
	return maps.Values(r.runners)
}

type Runner struct {
	Key        string
	Create     time.Time
	LastSeen   time.Time
	Endpoint   *url.URL
	Module     string
	Deployment model.DeploymentKey
}
