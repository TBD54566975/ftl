// Code generated by FTL. DO NOT EDIT.
package fsm

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type OpenDoorClient func(context.Context, Opened) error

type JamDoorClient func(context.Context, Jammed) error

type LockDoorClient func(context.Context, Locked) error

type SendEventClient func(context.Context, SendEventRequest) error

type UnlockDoorClient func(context.Context, Unlocked) error

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			OpenDoor,
		),
		reflection.ProvideResourcesForVerb(
			JamDoor,
		),
		reflection.ProvideResourcesForVerb(
			LockDoor,
		),
		reflection.ProvideResourcesForVerb(
			SendEvent,
		),
		reflection.ProvideResourcesForVerb(
			UnlockDoor,
		),
	)
}
