// Code generated by FTL. DO NOT EDIT.
package payment

import (
     "github.com/TBD54566975/ftl/go-runtime/ftl"
    "context"
    "github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
    ftlbuiltin "ftl/builtin"
)

type ChargeClient func(context.Context, ftlbuiltin.HttpRequest[ChargeRequest, ftl.Unit, ftl.Unit]) (ftlbuiltin.HttpResponse[ChargeResponse, ErrorResponse], error)

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
            Charge,
		),
	)
}