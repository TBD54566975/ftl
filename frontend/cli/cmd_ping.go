package main

import (
	"context"
	"time"

	"github.com/jpillora/backoff"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/rpc"
)

type pingCmd struct {
	Wait time.Duration `short:"w" help:"Wait up to this elapsed time for the FTL cluster to become available." default:"1s"`
}

func (c *pingCmd) Run(ctx context.Context, controller ftlv1connect.ControllerServiceClient) error {
	return rpc.Wait(ctx, backoff.Backoff{Max: time.Second}, c.Wait, controller) //nolint:wrapcheck
}
