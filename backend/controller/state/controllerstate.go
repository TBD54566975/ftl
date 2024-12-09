package state

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/eventstream"
	"github.com/TBD54566975/ftl/internal/model"
)

type State struct {
	runners             map[string]*Runner
	runnersByDeployment map[string][]*Runner
}

func NewInMemoryState() eventstream.EventStream[State] {
	return eventstream.NewInMemory(State{
		runners:             map[string]*Runner{},
		runnersByDeployment: map[string][]*Runner{},
	})
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
	Key        model.RunnerKey
	Create     time.Time
	LastSeen   time.Time
	Endpoint   string
	Module     string
	Deployment model.DeploymentKey
}

var _ eventstream.Event[State] = (*RunnerRegisteredEvent)(nil)
var _ eventstream.Event[State] = (*RunnerDeletedEvent)(nil)

type RunnerRegisteredEvent struct {
	Key        model.RunnerKey
	Time       time.Time
	Endpoint   string
	Module     string
	Deployment model.DeploymentKey
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

func (r RunnerDeletedEvent) Handle(t State) (State, error) {
	existing := t.runners[r.Key.String()]
	if existing != nil {
		delete(t.runners, r.Key.String())
	}
	return t, nil
}