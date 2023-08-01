package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type psCmd struct {
	All  bool `help:"Show all runners, even dead ones." short:"a"`
	JSON bool `help:"Output JSON."`
}

type process struct {
	Deployment string   `json:"deployment"`
	State      string   `json:"state"`
	Languages  []string `json:"language"`
	Module     string   `json:"module"`
	Runner     string   `json:"runner"`
}

func (s *psCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{
		AllRunners: s.All,
	}))
	if err != nil {
		return errors.WithStack(err)
	}
	var processes []process
	for _, runner := range status.Msg.Runners {
		if runner.State != ftlv1.RunnerState_RUNNER_ASSIGNED && !s.All {
			continue
		}
		deployment := deploymentByKey(status.Msg, runner.GetDeployment())
		module := ""
		if deployment != nil {
			module = deployment.Name
		}
		processes = append(processes, process{
			Deployment: runner.GetDeployment(),
			State:      strings.TrimPrefix(runner.State.String(), "RUNNER_"),
			Languages:  runner.Languages,
			Module:     module,
			Runner:     runner.Key,
		})
	}
	if s.JSON {
		for _, process := range processes {
			err = json.NewEncoder(os.Stdout).Encode(process)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}
	runnerFmt := "%-28s%-9s%-9s%-9s%-28s\n"
	fmt.Printf(runnerFmt, "Deployment", "State", "Language", "Module", "Runner")
	fmt.Printf(runnerFmt, strings.Repeat("-", 27), strings.Repeat("-", 8), strings.Repeat("-", 8), strings.Repeat("-", 8), strings.Repeat("-", 27))
	for _, process := range processes {
		if process.State != "ASSIGNED" && !s.All {
			continue
		}
		fmt.Printf(runnerFmt, process.Deployment, process.State, strings.Join(process.Languages, ":"), process.Module, process.Runner)
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
