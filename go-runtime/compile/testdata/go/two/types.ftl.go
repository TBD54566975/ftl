// Code generated by FTL. DO NOT EDIT.
package two

import (
	"context"
	ftlbuiltin "ftl/builtin"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	lib "github.com/TBD54566975/ftl/go-runtime/schema/testdata"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type CallsTwoClient func(context.Context, Payload[string]) (Payload[string], error)

type TwoClient func(context.Context, Payload[string]) (Payload[string], error)

type CallsTwoAndThreeClient func(context.Context, Payload[string]) (Payload[string], error)

type ThreeClient func(context.Context, Payload[string]) (Payload[string], error)

type CatchArrayClient func(context.Context, ftlbuiltin.CatchRequest[[]TwoEnum]) error

type ReturnsUserClient func(context.Context) (UserResponse, error)

func init() {
	reflection.Register(
		reflection.SumType[TypeEnum](
			*new(Exported),
			*new(List),
			*new(Scalar),
			*new(WithoutDirective),
		),
		reflection.ExternalType(*new(lib.NonFTLType)),
		reflection.ProvideResourcesForVerb(
			CallsTwo,
			server.VerbClient[TwoClient, Payload[string], Payload[string]](),
		),
		reflection.ProvideResourcesForVerb(
			Two,
		),
		reflection.ProvideResourcesForVerb(
			CallsTwoAndThree,
			server.VerbClient[TwoClient, Payload[string], Payload[string]](),
			server.VerbClient[ThreeClient, Payload[string], Payload[string]](),
		),
		reflection.ProvideResourcesForVerb(
			Three,
		),
		reflection.ProvideResourcesForVerb(
			CatchArray,
		),
		reflection.ProvideResourcesForVerb(
			ReturnsUser,
		),
	)
}