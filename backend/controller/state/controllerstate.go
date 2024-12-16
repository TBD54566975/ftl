package state

import (
	"github.com/block/ftl/internal/eventstream"
)

type State struct {
	deployments         map[string]*Deployment
	activeDeployments   map[string]*Deployment
	runners             map[string]*Runner
	runnersByDeployment map[string][]*Runner
	artifacts           map[string]bool
}

type ControllerEvent interface {
	Handle(view State) (State, error)
}

type ControllerState eventstream.EventStream[State, ControllerEvent]

func NewInMemoryState() ControllerState {
	return eventstream.NewInMemory[State, ControllerEvent](State{
		runners:             map[string]*Runner{},
		runnersByDeployment: map[string][]*Runner{},
	})
}
