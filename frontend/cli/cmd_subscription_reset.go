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
	// _, err := client.ResetSubscription(ctx, connect.NewRequest(&ftlpubsubv1.ResetSubscriptionRequest{
	// 	Subscription: s.Subscription.ToProto(),
	// }))
	// if err != nil {
	// 	return fmt.Errorf("failed to reset subscription: %w", err)
	// }
	// return nil
}
