package channels

import (
	"context"
	"iter"
)

// IterContext returns an iterator that iterates over the channel until the channel is closed or the context cancelled.
//
// Check ctx.Err() != nil to detect if the context was cancelled.
func IterContext[T any](ctx context.Context, ch <-chan T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				if !ok {
					return
				}
				if !yield(v) {
					return
				}
			}
		}
	}
}
