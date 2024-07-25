package ftltest

import (
	"context"
	"fmt"
	"reflect"

	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type fakeFSMInstance struct {
	name       string
	terminated bool
	state      reflect.Value
}

func newFakeFSMManager() *fakeFSMManager {
	return &fakeFSMManager{
		instances: make(map[fsmInstanceKey]*fakeFSMInstance),
	}
}

type fsmInstanceKey struct {
	fsm      string
	instance string
}

type fakeFSMManager struct {
	instances map[fsmInstanceKey]*fakeFSMInstance
}

func (f *fakeFSMManager) SendEvent(ctx context.Context, fsm string, instance string, event any) error {
	// Retrieve the FSM transitions.
	rfsm, ok := reflection.GetFSM(fsm).Get()
	if !ok {
		return fmt.Errorf("fsm %q not found", fsm)
	}
	schema := rfsm.Schema

	/// Upsert the FSM instance.
	key := fsmInstanceKey{fsm, instance}
	fsmInstance, ok := f.instances[key]
	if !ok {
		fsmInstance = &fakeFSMInstance{name: fsm}
		f.instances[key] = fsmInstance
	}

	if fsmInstance.terminated {
		return fmt.Errorf("fsm %q instance %q is terminated", fsm, instance)
	}

	// The function to execute.
	var transition reflection.Transition

	// Find the transition that matches the current state and the event type.
	for _, t := range rfsm.Transitions {
		if fsmInstance.state == t.From && reflect.TypeOf(event).AssignableTo(t.To.Type().In(1)) {
			transition = t
			break
		}
	}

	// Didn't find a transition.
	if !transition.To.IsValid() {
		return fmt.Errorf("no transition found for event %T", event)
	}

	out := transition.To.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(event)})
	var err error
	erri := out[0]
	if erri.IsNil() {
		fsmInstance.state = transition.To
	} else {
		err = erri.Interface().(error) //nolint:forcetypeassert
		fsmInstance.state = reflect.Value{}
	}
	currentStateRef := reflection.FuncRef(fsmInstance.state.Interface()).ToSchema()

	// Flag the FSM instance as terminated if the current state is a terminal state.
	for _, end := range schema.TerminalStates() {
		if currentStateRef.Equal(end) {
			fsmInstance.terminated = true
			break
		}
	}
	return err
}
