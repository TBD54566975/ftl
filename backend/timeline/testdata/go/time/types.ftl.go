// Code generated by FTL. DO NOT EDIT.
package time

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type InternalClient func(context.Context, TimeRequest) (TimeResponse, error)

type PublishInvoiceClient func(context.Context, PublishInvoiceRequest) error

type TimeClient func(context.Context, TimeRequest) (TimeResponse, error)

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			Internal,
		),
		reflection.ProvideResourcesForVerb(
			PublishInvoice,
			server.TopicHandle[Invoice, ftl.SinglePartitionMap[Invoice]]("time", "invoices"),
		),
		reflection.ProvideResourcesForVerb(
			Time,
		),
	)
}
