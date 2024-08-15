package fsmnext

import (
	"testing"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
	"github.com/alecthomas/assert/v2"
)

func TestProgression(t *testing.T) {
	// Simple progression through each state
	ctx := ftltest.Context()

	assert.NoError(t, SendOne(ctx, Request{
		State: A,
		Event: Event{
			Instance:   "1",
			NextStates: []State{B, C, D},
		},
	}))
}

func TestDoubleNext(t *testing.T) {
	// Bad progression where fsm.Next() is called twice
	ctx := ftltest.Context()

	assert.Contains(t, SendOne(ctx, Request{
		State: A,
		Event: Event{
			Instance:     "1",
			NextStates:   []State{B},
			NextAttempts: ftl.Some(2),
		},
	}).Error(), `fsm "fsm" instance "1" already has a pending event`)
}
