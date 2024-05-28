package fsm

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
)

func TestFSM(t *testing.T) {
	ctx := ftltest.Context()

	err := fsm.Send(ctx, "one", Two{Instance: "one"}) // No start transition on Two
	assert.Error(t, err)

	err = fsm.Send(ctx, "one", One{Instance: "one"}) // -> Start
	assert.NoError(t, err)
	err = fsm.Send(ctx, "one", One{Instance: "one"}) // -> Middle
	assert.NoError(t, err)
	err = fsm.Send(ctx, "one", One{Instance: "one"}) // -> End
	assert.NoError(t, err)

	err = fsm.Send(ctx, "one", One{Instance: "one"}) // Invalid, terminated
	assert.Error(t, err)
}
