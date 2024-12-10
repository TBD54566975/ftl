package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlpubsubv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	ftlv1pubsubconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/pubsub/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type resetSubscriptionCmd struct {
	Subscription reflection.Ref `arg:"" required:"" help:"Full path of subscription to reset."`
}

func (s *resetSubscriptionCmd) Run(ctx context.Context, client ftlv1pubsubconnect.LegacyPubsubServiceClient) error {
	_, err := client.ResetSubscription(ctx, connect.NewRequest(&ftlpubsubv1.ResetSubscriptionRequest{
		Subscription: s.Subscription.ToProto(),
	}))
	if err != nil {
		return fmt.Errorf("failed to reset subscription: %w", err)
	}
	return nil
}
