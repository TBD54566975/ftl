package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v2alpha1/v2alpha1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/v2/backend/routingservice"
	"github.com/TBD54566975/ftl/v2/backend/schemaservice"
)

type devCmd struct {
	Bind *url.URL `help:"The address to bind the schema service to." default:"http://127.0.0.1:9992"`
}

func (c *devCmd) Run(ctx context.Context, logger *log.Logger) error {
	logger = logger.Scope("dev")
	logger.Infof("Starting dev server on %s", c.Bind)
	schemaServiceClient := rpc.Dial(v2alpha1connect.NewSchemaServiceClient, c.Bind.String(), log.Error)
	err := rpc.Serve(ctx, c.Bind,
		rpc.GRPC(v2alpha1connect.NewSchemaServiceHandler, schemaservice.New()),
		rpc.GRPC(v2alpha1connect.NewRoutingServiceHandler, routingservice.New(ctx, schemaServiceClient)),
	)
	if err != nil {
		return fmt.Errorf("failed to start dev server: %w", err)
	}
	return nil
}
