// Code generated by FTL. DO NOT EDIT.
package time

import (
	"context"
	"github.com/TBD54566975/ftl/common/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type InternalClient func(context.Context, TimeRequest) (TimeResponse, error)

type TimeClient func(context.Context, TimeRequest) (TimeResponse, error)

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			Internal,
		),
		reflection.ProvideResourcesForVerb(
			Time,
			server.VerbClient[InternalClient, TimeRequest, TimeResponse](),
		),
	)
}
