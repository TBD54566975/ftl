// Code generated by FTL. DO NOT EDIT.
package origin

import (
	"context"
	ftlbuiltin "ftl/builtin"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type GetNonceClient func(context.Context, GetNonceRequest) (GetNonceResponse, error)

type PostAgentClient func(context.Context, ftlbuiltin.HttpRequest[Agent, ftl.Unit, ftl.Unit]) (ftlbuiltin.HttpResponse[PostAgentResponse, PostAgentErrorResponse], error)

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			GetNonce,
			server.Config[string]("origin", "nonce"),
		),
		reflection.ProvideResourcesForVerb(
			PostAgent,
			server.TopicHandle[Agent, ftl.SinglePartitionMap[Agent]]("origin", "agentBroadcast"),
		),
	)
}
