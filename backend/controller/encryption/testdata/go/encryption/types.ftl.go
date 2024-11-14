// Code generated by FTL. DO NOT EDIT.
package encryption

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type ConsumeClient func(context.Context, Event) error

type EchoClient func(context.Context, EchoRequest) (EchoResponse, error)

type PublishClient func(context.Context, Event) error

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			Consume,
		),
		reflection.ProvideResourcesForVerb(
			Echo,
		),
		reflection.ProvideResourcesForVerb(
			Publish,
			server.TopicHandle[Event]("encryption", "topic"),
		),
	)
}
