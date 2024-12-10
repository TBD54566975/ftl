package state

import (
	"github.com/TBD54566975/ftl/internal/eventstream"
)

type State struct {
	deployments         map[string]*Deployment
	activeDeployments   map[string]*Deployment
	runners             map[string]*Runner
	runnersByDeployment map[string][]*Runner
	artifacts           map[string]bool
}

func NewInMemoryState() eventstream.EventStream[State] {
	return eventstream.NewInMemory(State{
		runners:             map[string]*Runner{},
		runnersByDeployment: map[string][]*Runner{},
	})
}
