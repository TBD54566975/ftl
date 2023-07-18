package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/golang/protobuf/jsonpb"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type psCmd struct {
	All  bool `help:"Show all runners, even dead ones." short:"a"`
	JSON bool `help:"Output JSON."`
}

func (s *psCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{
		AllRunners: s.All,
	}))
	if err != nil {
		return errors.WithStack(err)
	}
	if s.JSON {
		fmt.Print("[")
		for i, runner := range status.Msg.Runners {
			if i != 0 {
				fmt.Print(",")
			}
			err = errors.WithStack((&jsonpb.Marshaler{}).Marshal(os.Stdout, runner))
			if err != nil {
				return errors.WithStack(err)
			}
		}
		fmt.Print("]")
		return nil
	}
	runnerFmt := "%-28s%-9s%-9s%-9s%-28s\n"
	fmt.Printf(runnerFmt, "Deployment", "State", "Language", "Module", "Runner")
	fmt.Printf(runnerFmt, strings.Repeat("-", 27), strings.Repeat("-", 8), strings.Repeat("-", 8), strings.Repeat("-", 8), strings.Repeat("-", 27))
	for _, runner := range status.Msg.Runners {
		if runner.State != ftlv1.RunnerState_RUNNER_ASSIGNED && !s.All {
			continue
		}
		deployment := deploymentByKey(status.Msg, runner.GetDeployment())
		module := ""
		if deployment != nil {
			module = deployment.Name
		}
		fmt.Printf(runnerFmt, runner.GetDeployment(), strings.TrimPrefix(runner.State.String(), "RUNNER_"), runner.Language, module, runner.Key)
	}
	return nil
}

func deploymentByKey(resp *ftlv1.StatusResponse, key string) *ftlv1.StatusResponse_Deployment {
	for _, deployment := range resp.Deployments {
		if deployment.Key == key {
			return deployment
		}
	}
	return nil
}
