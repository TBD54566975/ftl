package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	jsonpb "google.golang.org/protobuf/encoding/protojson"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
)

type statusCmd struct {
	All              bool `help:"Show all controllers, deployments, and runners, even those that are not running."`
	AllControllers   bool `help:"Show all controllers, even those that are not running."`
	AllRunners       bool `help:"Show all runners, even those that are not running."`
	AllIngressRoutes bool `help:"Show all ingress routes, even those that are not running."`
	Schema           bool `help:"Show schema."`
}

func (s *statusCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{}))
	if err != nil {
		return err
	}
	msg := status.Msg
	if !s.Schema {
		for _, deployment := range msg.Deployments {
			deployment.Schema = nil
		}
	}
	marshaler := jsonpb.MarshalOptions{Indent: "  "}
	data, err := marshaler.Marshal(status.Msg)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}
	fmt.Printf("%s\n", data)
	return nil
}
