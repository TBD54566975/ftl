package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestFSM(t *testing.T) {
	f := &FSM{
		Name:  "payment",
		Start: []*Ref{{Name: "created"}, {Name: "paid"}},
		Transitions: []*FSMTransition{
			{From: &Ref{Name: "created"}, To: &Ref{Name: "paid"}},
			{From: &Ref{Name: "created"}, To: &Ref{Name: "failed"}},
			{From: &Ref{Name: "paid"}, To: &Ref{Name: "completed"}},
		},
	}
	assert.Equal(t, []*Ref{{Name: "completed"}, {Name: "failed"}}, f.TerminalStates())

	f = &FSM{
		Name:  "fsm",
		Start: []*Ref{{Name: "start"}},
		Transitions: []*FSMTransition{
			{From: &Ref{Name: "start"}, To: &Ref{Name: "middle"}},
			{From: &Ref{Name: "middle"}, To: &Ref{Name: "end"}},
		},
	}
	assert.Equal(t, []*Ref{{Name: "end"}}, f.TerminalStates())
}
