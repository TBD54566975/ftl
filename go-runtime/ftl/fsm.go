package ftl

import "github.com/TBD54566975/ftl/go-runtime/ftl/reflection"

type FSMHandle struct {
	transitions []FSMTransition
}

type FSMTransition struct {
	from reflection.Ref
	to   reflection.Ref
}

// Start specifies a start state in an FSM.
func Start[In any](state Sink[In]) FSMTransition {
	return FSMTransition{to: reflection.FuncRef(state)}
}

// Transition specifies a transition in an FSM.
//
// The "event" triggering the transition is the input to the "from" state.
func Transition[FromIn, ToIn any](from Sink[FromIn], to Sink[ToIn]) FSMTransition {
	return FSMTransition{from: reflection.FuncRef(from), to: reflection.FuncRef(to)}
}

// FSM creates a new finite-state machine.
func FSM(name string, transitions ...FSMTransition) *FSMHandle {
	return &FSMHandle{transitions: transitions}
}
