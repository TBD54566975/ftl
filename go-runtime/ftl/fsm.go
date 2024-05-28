package ftl

import (
	"context"
	"reflect"

	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
)

type FSMHandle struct {
	name string
}

type FSMTransition struct {
	fromFunc reflect.Value
	from     reflection.Ref
	toFunc   reflect.Value
	to       reflection.Ref
}

// Start specifies a start state in an FSM.
func Start[In any](state Sink[In]) FSMTransition {
	return FSMTransition{
		toFunc: reflect.ValueOf(state),
		to:     reflection.FuncRef(state),
	}
}

// Transition specifies a transition in an FSM.
//
// The "event" triggering the transition is the input to the "from" state.
func Transition[FromIn, ToIn any](from Sink[FromIn], to Sink[ToIn]) FSMTransition {
	return FSMTransition{
		fromFunc: reflect.ValueOf(from),
		from:     reflection.FuncRef(from),
		toFunc:   reflect.ValueOf(to),
		to:       reflection.FuncRef(to),
	}
}

// FSM creates a new finite-state machine.
func FSM(name string, transitions ...FSMTransition) *FSMHandle {
	rtransitions := make([]reflection.Transition, len(transitions))
	for i, transition := range transitions {
		rtransitions[i] = reflection.Transition{From: transition.fromFunc, To: transition.toFunc}
	}
	reflection.Register(reflection.FSM(name, rtransitions...))
	return &FSMHandle{name: name}
}

// Send an event to an instance of the FSM.
//
// "instance" must uniquely identify an instance of the FSM. The event type must
// be valid for the current state of the FSM instance.
//
// If the FSM instance is not executing, a new one will be started. If the event
// is not valid for the current state, an error will be returned.
func (f *FSMHandle) Send(ctx context.Context, instance string, event any) error {
	return internal.FromContext(ctx).FSMSend(ctx, f.name, instance, event)
}
