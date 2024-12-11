// Code generated by FTL. DO NOT EDIT.
package slow

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/common/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type ConsumeClient func(context.Context, Event) error

type PublishClient func(context.Context, PublishRequest) error

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			Consume,
		),
		reflection.ProvideResourcesForVerb(
			Publish,
			server.TopicHandle[Event, ftl.SinglePartitionMap[Event]]("slow", "topic"),
		),
	)
}
