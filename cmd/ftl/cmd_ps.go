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
	runnerFmt := "%-28s%-9s%-9s%-28s%s\n"
	fmt.Printf(runnerFmt, "Runner", "State", "Language", "Deployment", "Endpoint")
	fmt.Printf(runnerFmt, strings.Repeat("-", 27), strings.Repeat("-", 8), strings.Repeat("-", 8), strings.Repeat("-", 27), strings.Repeat("-", 8))
	for _, runner := range status.Msg.Runners {
		fmt.Printf(runnerFmt, runner.Key, strings.TrimPrefix(runner.State.String(), "RUNNER_"), runner.Language, runner.GetDeployment(), runner.Endpoint)
	}
	fmt.Println()
	return nil
}
