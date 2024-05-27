package ftl

import (
	"context"
	"fmt"
	"reflect"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type FSMHandle struct {
	name        string
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
	return &FSMHandle{name: name, transitions: transitions}
}

// Send an event to an instance of the FSM.
//
// "instance" must uniquely identify an instance of the FSM. The event type must
// be valid for the current state of the FSM instance.
//
// If the FSM instance is not executing, a new one will be started. If the event
// is not valid for the current state, an error will be returned.
func (f *FSMHandle) Send(ctx context.Context, instance string, event any) error {
	client := rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
	body, err := encoding.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	_, err = client.SendFSMEvent(ctx, connect.NewRequest(&ftlv1.SendFSMEventRequest{
		Fsm:      &schemapb.Ref{Module: reflection.Module(), Name: f.name},
		Instance: instance,
		Event:    schema.TypeToProto(reflection.ReflectTypeToSchemaType(reflect.TypeOf(event))),
		Body:     body,
	}))
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}
	return nil
}
