package fsmretry

import (
	"context"
	"fmt"
	"ftl/builtin"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:retry 2 2s 3s
var fsm = ftl.FSM("fsm",
	ftl.Start(State1),
	ftl.Transition(State1, State2),
	ftl.Transition(State1, State3),
)

type StartEvent struct {
	ID string `json:"id"`
}

type TransitionToTwoEvent struct {
	ID        string `json:"id"`
	FailCatch bool
}

type TransitionToThreeEvent struct {
	ID string `json:"id"`
}

//ftl:verb
func State1(ctx context.Context, in StartEvent) error {
	return nil
}

//ftl:verb
//ftl:retry 2 2s 2s catch catchState2
func State2(ctx context.Context, in TransitionToTwoEvent) error {
	return fmt.Errorf("transition will never succeed")
}

//ftl:verb
func CatchState2(ctx context.Context, in builtin.CatchRequest[TransitionToTwoEvent]) error {
	if in.Request.FailCatch {
		return fmt.Errorf("catching failed as requested by event")
	}
	return nil
}

// State3 will have its retry policy defaulted to the fsm one
//
//ftl:verb
func State3(ctx context.Context, in TransitionToThreeEvent) error {
	return fmt.Errorf("transition will never succeed")
}

//ftl:verb
func Start(ctx context.Context, in StartEvent) error {
	return fsm.Send(ctx, in.ID, in)
}

//ftl:verb
func StartTransitionToTwo(ctx context.Context, in TransitionToTwoEvent) error {
	return fsm.Send(ctx, in.ID, in)
}

//ftl:verb
func StartTransitionToThree(ctx context.Context, in TransitionToThreeEvent) error {
	return fsm.Send(ctx, in.ID, in)
}
