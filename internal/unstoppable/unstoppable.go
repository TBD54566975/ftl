package unstoppable

import (
	"context"
	"time"
)

//nolint:containedctx
type unstoppableContext struct {
	parent context.Context
}

func (u unstoppableContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (u unstoppableContext) Done() <-chan struct{} {
	return nil
}

func (u unstoppableContext) Err() error {
	return nil
}

func (u unstoppableContext) Value(key any) any {
	return u.parent.Value(key)
}

// Context returns a sub-context that is not cancellable.
func Context(ctx context.Context) context.Context {
	return unstoppableContext{parent: ctx}
}
