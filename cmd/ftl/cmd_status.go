package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/internal/slices"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/schema"
)

type statusCmd struct {
	Schema bool `help:"Show schema."`
}

func (s *statusCmd) Run(ctx context.Context, client ftlv1connect.ControlPlaneServiceClient) error {
	status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{}))
	if err != nil {
		return errors.WithStack(err)
	}
	runnerFmt := "%-27s%-9s%-9s%-27s%s\n"
	fmt.Printf(runnerFmt, "Runner", "State", "Language", "Deployment", "Endpoint")
	fmt.Printf(runnerFmt, strings.Repeat("-", 26), strings.Repeat("-", 8), strings.Repeat("-", 8), strings.Repeat("-", 26), strings.Repeat("-", 8))
	for _, runner := range status.Msg.Runners {
		fmt.Printf(runnerFmt, cleanKey(runner.Key), runner.State, runner.Language, cleanKey(runner.GetDeployment()), runner.Endpoint)
	}
	fmt.Println()
	deploymentFmt := "%-27s%-15s%-5v%s\n"
	if s.Schema {
		fmt.Printf(deploymentFmt, "Deployment", "Name", "N", "Schema")
		fmt.Printf(deploymentFmt, strings.Repeat("-", 26), strings.Repeat("-", 14), strings.Repeat("-", 4), strings.Repeat("-", 33))
	} else {
		fmt.Printf(deploymentFmt, "Deployment", "Name", "N", "")
		fmt.Printf(deploymentFmt, strings.Repeat("-", 26), strings.Repeat("-", 14), strings.Repeat("-", 4), "")
	}
	for _, deployment := range status.Msg.Deployments {
		active := slices.Reduce(status.Msg.Runners, 0, func(i int, runner *ftlv1.StatusResponse_Runner) int {
			if runner.GetDeployment() == deployment.Key {
				return i + 1
			}
			return i
		})
		count := fmt.Sprintf("%d/%d", active, deployment.MinReplicas)
		if s.Schema {
			ms, err := schema.ModuleFromProto(deployment.Schema)
			if err != nil {
				return errors.Wrapf(err, "%q: invalid schema", deployment.Name)
			}
			mst := indent(ms.String(), 47)
			fmt.Printf(deploymentFmt, cleanKey(deployment.Key), deployment.Name, count, mst)
		} else {
			fmt.Printf(deploymentFmt, cleanKey(deployment.Key), deployment.Name, count, "")
		}
	}
	return nil
}

// Indent every line in "s" by "indent".
func indent(s string, indent int) string {
	return strings.ReplaceAll(s, "\n", "\n"+strings.Repeat(" ", indent))
}

func cleanKey(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) == 3 {
		return parts[2]
	}
	return key
}
