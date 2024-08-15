package fsmnext

import (
	"context"
	"errors"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

func fsm() *ftl.FSMHandle {
	// This FSM allows transitions moving forward through the alphabet
	// Each transition also declares the next state(s) to transition to using State
	//
	//ftl:retry 2 1s
	var fsm = ftl.FSM("fsm",
		ftl.Start(StateA),
		ftl.Transition(StateA, StateB),
		ftl.Transition(StateA, StateC),
		ftl.Transition(StateA, StateD),
		ftl.Transition(StateB, StateC),
		ftl.Transition(StateB, StateD),
		ftl.Transition(StateC, StateD),
	)
	return fsm
}

type State string

const (
	A State = "A"
	B State = "B"
	C State = "C"
	D State = "D"
)

type Event struct {
	Instance     string
	NextStates   []State            // will schedule fsm.Next with these states progressively
	NextAttempts ftl.Option[int]    // will call fsm.Next this many times. Once otherwise
	Error        ftl.Option[string] // if present, returns this error after calling fsm.Next() as needed
}

//ftl:typealias
type EventA Event

//ftl:verb
func StateA(ctx context.Context, in EventA) error {
	return handleEvent(ctx, Event(in))
}

//ftl:typealias
type EventB Event

//ftl:verb
func StateB(ctx context.Context, in EventB) error {
	return handleEvent(ctx, Event(in))
}

//ftl:typealias
type EventC Event

//ftl:verb
func StateC(ctx context.Context, in EventC) error {
	return handleEvent(ctx, Event(in))
}

//ftl:typealias
type EventD Event

//ftl:verb
func StateD(ctx context.Context, in EventD) error {
	return handleEvent(ctx, Event(in))
}

//ftl:data export
type Request struct {
	State State
	Event Event
}

//ftl:verb export
func SendOne(ctx context.Context, in Request) error {
	return fsm().Send(ctx, in.Event.Instance, eventFor(in.Event, in.State))
}

func handleEvent(ctx context.Context, in Event) error {
	if len(in.NextStates) == 0 {
		return nil
	}
	event := eventFor(Event{
		Instance:     in.Instance,
		NextStates:   in.NextStates[1:],
		NextAttempts: in.NextAttempts,
	}, in.NextStates[0])
	attempts := in.NextAttempts.Default(1)
	for i := range attempts {
		ftl.LoggerFromContext(ctx).Infof("scheduling next event for %s (%d/%d)", in.Instance, i+1, attempts)
		if err := fsm().Next(ctx, in.Instance, event); err != nil {
			return err
		}
	}
	if errStr, ok := in.Error.Get(); ok {
		return errors.New(errStr)
	}
	return nil
}

func eventFor(event Event, state State) any {
	switch state {
	case A:
		return EventA(event)
	case B:
		return EventB(event)
	case C:
		return EventC(event)
	case D:
		return EventD(event)
	default:
		panic("unknown state")
	}
}
