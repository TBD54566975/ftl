package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/golang/protobuf/jsonpb"
	"golang.org/x/exp/maps"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type psCmd struct {
	Verbose int  `help:"Show process detail." short:"v" type:"counter"`
	JSON    bool `help:"Output JSON."`
}

func (s *psCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	status, err := client.ProcessList(ctx, connect.NewRequest(&ftlv1.ProcessListRequest{}))
	if err != nil {
		return errors.WithStack(err)
	}
	if s.JSON {
		marshaller := jsonpb.Marshaler{Indent: "  "}
		for _, process := range status.Msg.Processes {
			err = marshaller.Marshal(os.Stdout, process)
			if err != nil {
				return errors.WithStack(err)
			}
			fmt.Println()
		}
		return nil
	}

	groups := map[string][]*ftlv1.ProcessListResponse_Process{}
	for _, process := range status.Msg.Processes {
		groups[process.Deployment] = append(groups[process.Deployment], process)
	}

	keys := maps.Keys(groups)
	sort.Strings(keys)

	if s.Verbose > 0 {
		format := "%-40s %-10s %-10s %-27s %-50s"
		columns := []any{"DEPLOYMENT", "REPLICAS", "STATE", "RUNNER", "ENDPOINT"}
		if s.Verbose > 1 {
			format += " %-20s %-20s"
			columns = append(columns, "DEPLOYMENT-LABELS", "RUNNER-LABELS")
		}
		format += "\n"
		fmt.Printf(format, columns...)
		for _, key := range keys {
			first := groups[key][0]
			for i, process := range groups[key] {
				key := key
				runnerKey := "-"
				endpoint := "-"
				runnerLabels := "-"
				if runner := process.Runner; runner != nil {
					key = fmt.Sprintf("%s-%s", key, strings.ToLower(runner.Key[len(runner.Key)-5:]))
					runnerKey = runner.Key
					endpoint = runner.Endpoint
					labels, err := (&jsonpb.Marshaler{}).MarshalToString(runner.Labels)
					if err != nil {
						return errors.WithStack(err)
					}
					runnerLabels = labels
				}
				args := []any{key, fmt.Sprintf("%d/%d", i+1, first.MinReplicas), "live", runnerKey, endpoint}
				if s.Verbose > 1 {
					labels, err := (&jsonpb.Marshaler{}).MarshalToString(first.Labels)
					if err != nil {
						return errors.WithStack(err)
					}
					args = append(args, labels, runnerLabels)
				}
				fmt.Printf(format, args...)
			}
			for i := len(groups[key]); i < int(first.MinReplicas); i++ {
				args := []any{key + "-?????", fmt.Sprintf("%d/%d", i+1, first.MinReplicas), "-", "-", "-"}
				if s.Verbose > 1 {
					args = append(args, "-", "-")
				}
				fmt.Printf(format, args...)
			}
		}
	} else {
		// Format: deployment, N/M replicas running
		format := "%-30s %d/%d\n"
		fmt.Printf("%-30s %s\n", "DEPLOYMENT", "REPLICAS")
		for _, key := range keys {
			fmt.Printf(format, key, len(groups[key]), groups[key][0].MinReplicas)
		}
	}
	return nil
}
