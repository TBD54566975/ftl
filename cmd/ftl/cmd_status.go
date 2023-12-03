package main

import (
	"context"
	"os"

	"connectrpc.com/connect"
	"github.com/golang/protobuf/jsonpb"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type statusCmd struct {
	All              bool `help:"Show all controllers, deployments, and runners, even those that are not running."`
	AllControllers   bool `help:"Show all controllers, even those that are not running."`
	AllDeployments   bool `help:"Show all deployments, even those that are not running."`
	AllRunners       bool `help:"Show all runners, even those that are not running."`
	AllIngressRoutes bool `help:"Show all ingress routes, even those that are not running."`
	Schema           bool `help:"Show schema."`
}

func (s *statusCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{
		AllControllers:   s.All || s.AllControllers,
		AllDeployments:   s.All || s.AllDeployments,
		AllRunners:       s.All || s.AllRunners,
		AllIngressRoutes: s.All || s.AllIngressRoutes,
	}))
	if err != nil {
		return err
	}
	msg := status.Msg
	if !s.Schema {
		for _, deployment := range msg.Deployments {
			deployment.Schema = nil
		}
	}
	return (&jsonpb.Marshaler{
		Indent: "  ",
	}).Marshal(os.Stdout, status.Msg)
}
