package unstoppable

import (
	"context"
	"time"
)

// Context returns a sub-context that is not cancellable.
func Context(ctx context.Context) context.Context {
	return unstoppableContext{parent: ctx, ch: make(chan struct{})}
}

//nolint:containedctx
type unstoppableContext struct {
	parent context.Context
	ch     chan struct{}
}

func (u unstoppableContext) Deadline() (deadline time.Time, ok bool) { return time.Time{}, false }
func (u unstoppableContext) Done() <-chan struct{}                   { return u.ch }
func (u unstoppableContext) Err() error                              { return nil }
func (u unstoppableContext) Value(key any) any                       { return u.parent.Value(key) }
