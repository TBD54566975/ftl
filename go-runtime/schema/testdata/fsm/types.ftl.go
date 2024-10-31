// Code generated by FTL. DO NOT EDIT.
package fsm

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type CreatedClient func(context.Context, OnlinePaymentCreated) error

type PaidClient func(context.Context, OnlinePaymentPaid) error

type CompletedClient func(context.Context, OnlinePaymentCompleted) error

type FailedClient func(context.Context, OnlinePaymentFailed) error

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			Created,
		),
		reflection.ProvideResourcesForVerb(
			Paid,
		),
		reflection.ProvideResourcesForVerb(
			Completed,
		),
		reflection.ProvideResourcesForVerb(
			Failed,
		),
	)
}