// Code generated by FTL. DO NOT EDIT.
package cron

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type HourlyClient func(context.Context) error

type ThirtySecondsClient func(context.Context) error

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			Hourly,
		),
		reflection.ProvideResourcesForVerb(
			ThirtySeconds,
		),
	)
}