package ftltest

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/internal"
)

type fakeFTL struct {
	fsm *fakeFSMManager
}

func newFakeFTL() *fakeFTL {
	return &fakeFTL{
		fsm: newFakeFSMManager(),
	}
}

var _ internal.FTL = &fakeFTL{}

func (f *fakeFTL) FSMSend(ctx context.Context, fsm string, instance string, event any) error {
	return f.fsm.SendEvent(ctx, fsm, instance, event)
}
