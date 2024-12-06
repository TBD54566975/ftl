package controller

import (
	"fmt"
	"net/url"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/eventstream"
	"github.com/TBD54566975/ftl/internal/model"
)

type State struct {
	runners             map[string]*Runner
	runnersByDeployment map[string][]*Runner
}

func (r *State) Runner(s string) optional.Option[Runner] {
	result, ok := r.runners[s]
	if ok {
		return optional.Ptr(result)
	}
	return optional.None[Runner]()
}

func (r *State) Runners() []Runner {
	var ret []Runner
	for _, v := range r.runners {
		ret = append(ret, *v)
	}
	return ret
}

func (r *State) RunnersForDeployment(deployment string) []Runner {
	var ret []Runner
	for _, v := range r.runnersByDeployment[deployment] {
		ret = append(ret, *v)
	}
	return ret
}

type Runner struct {
	Key        string
	Create     time.Time
	LastSeen   time.Time
	Endpoint   *url.URL
	Module     string
	Deployment model.DeploymentKey
}

var _ eventstream.Event[State] = (*RunnerCreatedEvent)(nil)
var _ eventstream.Event[State] = (*RunnerHeartbeatEvent)(nil)
var _ eventstream.Event[State] = (*RunnerDeletedEvent)(nil)

type RunnerCreatedEvent struct {
	Key        string
	Create     time.Time
	Endpoint   *url.URL
	Module     string
	Deployment model.DeploymentKey
}

type RunnerHeartbeatEvent struct {
	Key      string
	LastSeen time.Time
}

func (r *RunnerHeartbeatEvent) Handle(t State) (State, error) {
	existing := t.runners[r.Key]
	if existing == nil {
		return t, fmt.Errorf("runner %s not found", r.Key)
	}
	existing.LastSeen = r.LastSeen
	return t, nil
}

type RunnerDeletedEvent struct {
	Key string
}

func (r RunnerDeletedEvent) Handle(t State) (State, error) {
	existing := t.runners[r.Key]
	if existing != nil {
		delete(t.runners, r.Key)

	}
	return t, nil
}

func (r *RunnerCreatedEvent) Handle(t State) (State, error) {
	n := Runner{
		Key:        r.Key,
		Create:     r.Create,
		Endpoint:   r.Endpoint,
		Module:     r.Module,
		Deployment: r.Deployment,
	}
	t.runners[r.Key] = &n
	t.runnersByDeployment[r.Deployment.String()] = append(t.runnersByDeployment[r.Deployment.String()], &n)
	return t, nil
}
