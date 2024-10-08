package fsm

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type Opened struct{}
type Unlocked struct{}
type Locked struct{}
type Jammed struct{}

var door = ftl.FSM(
	"door",
	ftl.Start(OpenDoor),
	ftl.Transition(OpenDoor, UnlockDoor),
	ftl.Transition(OpenDoor, JamDoor),
	ftl.Transition(UnlockDoor, OpenDoor),
	ftl.Transition(UnlockDoor, LockDoor),
	ftl.Transition(UnlockDoor, JamDoor),
	ftl.Transition(LockDoor, UnlockDoor),
	ftl.Transition(LockDoor, JamDoor),
	ftl.Transition(JamDoor, OpenDoor),
)

//ftl:verb
func OpenDoor(ctx context.Context, in Opened) error {
	fmt.Println("The door is open.")
	return nil
}

//ftl:verb
func UnlockDoor(ctx context.Context, in Unlocked) error {
	fmt.Println("The door is unlocked.")
	return nil
}

//ftl:verb
func LockDoor(ctx context.Context, in Locked) error {
	fmt.Println("The door is locked.")
	return nil
}

//ftl:verb
func JamDoor(ctx context.Context, in Jammed) error {
	fmt.Println("The door is jammed. Fixing...")
	ftl.FSMNext(ctx, Opened{})
	return nil
}

//ftl:enum
type Event string

const (
	Open   Event = "open"
	Unlock Event = "unlock"
	Lock   Event = "lock"
	Jam    Event = "jam"
)

type SendEventRequest struct {
	ID    string
	Event Event
}

//ftl:verb export
func SendEvent(ctx context.Context, req SendEventRequest) error {
	switch req.Event {
	case Open:
		return door.Send(ctx, req.ID, Opened{})
	case Unlock:
		return door.Send(ctx, req.ID, Unlocked{})
	case Lock:
		return door.Send(ctx, req.ID, Locked{})
	case Jam:
		return door.Send(ctx, req.ID, Jammed{})
	default:
		return fmt.Errorf("unknown event: %s", req.Event)
	}
}
