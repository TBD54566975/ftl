package main

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type resetSubscriptionCmd struct {
	Subscription reflection.Ref `arg:"" required:"" help:"Full path of subscription to reset."`
}

func (s *resetSubscriptionCmd) Run(ctx context.Context) error {
	panic("not implemented")
}
