package internal

import (
	"context"
)

// FTL is the interface that the FTL runtime provides to user code.
//
// In production, the FTL runtime will provide an implementation of this
// interface that communicates with the Controller over gRPC.
//
// In testing code, the implementation will inject fakes and other test
// implementations.
type FTL interface {
	// FSMSend sends an event to an instance of an FSM.
	FSMSend(ctx context.Context, fsm, instance string, data any) error

	// PublishEvent sends an event to a pubsub topic.
	PublishEvent(ctx context.Context, topic string, event any) error

	// CallMap calls Get on an instance of an ftl.Map.
	CallMap(ctx context.Context, mapper any, mapImpl func(context.Context) (any, error)) any
}

type ftlContextKey struct{}

// WithContext returns a new context with the FTL instance.
func WithContext(ctx context.Context, ftl FTL) context.Context {
	return context.WithValue(ctx, ftlContextKey{}, ftl)
}

// FromContext returns the FTL instance from the context.
func FromContext(ctx context.Context) FTL {
	ftl, ok := ctx.Value(ftlContextKey{}).(FTL)
	if !ok {
		panic("FTL not found in context")
	}
	return ftl
}
