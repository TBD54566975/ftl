package ftl

type FSMHandle struct {
	transitions []FSMTransition
}

type FSMTransition struct {
	from Ref
	to   Ref
}

// Start specifies a start state in an FSM.
func Start[In any](state Sink[In]) FSMTransition {
	return FSMTransition{to: FuncRef(state)}
}

// Transition specifies a transition in an FSM.
//
// The "event" triggering the transition is the input to the "from" state.
func Transition[FromIn, ToIn any](from Sink[FromIn], to Sink[ToIn]) FSMTransition {
	return FSMTransition{from: FuncRef(from), to: FuncRef(to)}
}

// FSM creates a new finite-state machine.
func FSM(name string, transitions ...FSMTransition) *FSMHandle {
	return &FSMHandle{transitions: transitions}
}
