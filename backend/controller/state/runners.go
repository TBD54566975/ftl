package state

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/eventstream"
	"github.com/block/ftl/internal/model"
)

type Runner struct {
	Key        model.RunnerKey
	Create     time.Time
	LastSeen   time.Time
	Endpoint   string
	Module     string
	Deployment model.DeploymentKey
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

var _ ControllerEvent = (*RunnerRegisteredEvent)(nil)
var _ eventstream.VerboseMessage = (*RunnerRegisteredEvent)(nil)
var _ ControllerEvent = (*RunnerDeletedEvent)(nil)

type RunnerRegisteredEvent struct {
	Key        model.RunnerKey
	Time       time.Time
	Endpoint   string
	Module     string
	Deployment model.DeploymentKey
}

func (r *RunnerRegisteredEvent) VerboseMessage() {
	// Stops this message being logged every second
}

func (r *RunnerRegisteredEvent) Handle(t State) (State, error) {
	if existing := t.runners[r.Key.String()]; existing != nil {
		existing.LastSeen = r.Time
		return t, nil
	}
	n := Runner{
		Key:        r.Key,
		Create:     r.Time,
		LastSeen:   r.Time,
		Endpoint:   r.Endpoint,
		Module:     r.Module,
		Deployment: r.Deployment,
	}
	t.runners[r.Key.String()] = &n
	t.runnersByDeployment[r.Deployment.String()] = append(t.runnersByDeployment[r.Deployment.String()], &n)
	return t, nil
}

type RunnerDeletedEvent struct {
	Key model.RunnerKey
}

func (r *RunnerDeletedEvent) Handle(t State) (State, error) {
	existing := t.runners[r.Key.String()]
	if existing != nil {
		delete(t.runners, r.Key.String())
	}
	return t, nil
}
