package ftltest

import (
	"context"
	"fmt"
	"reflect"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
)

type fakeFSMInstance struct {
	name       string
	terminated bool
	state      reflect.Value
	next       ftl.Option[any]
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
		return fmt.Errorf(`invalid event "%T" for state "%v"`, event, fsmInstance.state.Type().In(1))
	}

	callCtx := internal.ContextWithCallMetadata(ctx, map[string]string{
		"fsmName":     fsm,
		"fsmInstance": instance,
	})
	out := transition.To.Call([]reflect.Value{reflect.ValueOf(callCtx), reflect.ValueOf(event)})
	erri := out[0]
	if !erri.IsNil() {
		err := erri.Interface().(error) //nolint:forcetypeassert
		fsmInstance.state = reflect.Value{}
		fsmInstance.next = ftl.None[any]()
		fsmInstance.terminated = true
		return err
	}

	fsmInstance.state = transition.To

	currentStateRef := reflection.FuncRef(fsmInstance.state.Interface()).ToSchema()

	// Flag the FSM instance as terminated if the current state is a terminal state.
	for _, end := range schema.TerminalStates() {
		if currentStateRef.Equal(end) {
			fsmInstance.terminated = true
			break
		}
	}

	if next, ok := fsmInstance.next.Get(); ok {
		fsmInstance.next = ftl.None[any]()
		return f.SendEvent(ctx, fsm, instance, next)
	}
	return nil
}

func (f *fakeFSMManager) SetNextFSMEvent(ctx context.Context, fsm string, instance string, event any) error {
	key := fsmInstanceKey{fsm, instance}
	fsmInstance, ok := f.instances[key]
	if !ok {
		return fmt.Errorf("fsm %q instance %q not found", fsm, instance)
	}
	if fsmInstance.terminated {
		return fmt.Errorf("fsm %q instance %q is terminated", fsm, instance)
	}
	if _, ok := fsmInstance.next.Get(); ok {
		return fmt.Errorf("fsm %q instance %q already has a pending event", fsm, instance)
	}
	fsmInstance.next = ftl.Some(event)
	return nil
}
