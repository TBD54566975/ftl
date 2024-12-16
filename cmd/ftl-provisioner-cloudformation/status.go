package main

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	provisioner "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
)

func (c *CloudformationProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	token := req.Msg.ProvisioningToken
	// if the task is not in the map, it means that the provisioner has crashed since starting the task
	// in that case, we start a new task to query the existing stack
	task, loaded := c.running.LoadOrStore(token, &task{stackID: token})
	if !loaded {
		task.Start(ctx, c.client, c.secrets, "")
	}

	if task.err.Load() != nil {
		c.running.Delete(token)
		return nil, connect.NewError(connect.CodeUnknown, task.err.Load())
	}

	if task.outputs.Load() != nil {
		c.running.Delete(token)

		events, err := c.updateResources(ctx, task.outputs.Load())
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&provisioner.StatusResponse{
			Status: &provisioner.StatusResponse_Success{
				Success: &provisioner.StatusResponse_ProvisioningSuccess{
					Events: events,
				},
			},
		}), nil
	}

	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.StatusResponse_Running{
			Running: &provisioner.StatusResponse_ProvisioningRunning{},
		},
	}), nil
}

func outputsByResourceID(outputs []types.Output) (map[string][]types.Output, error) {
	m := make(map[string][]types.Output)
	for _, output := range outputs {
		key, err := decodeOutputKey(output)
		if err != nil {
			return nil, fmt.Errorf("failed to decode output key: %w", err)
		}
		m[key.ResourceID] = append(m[key.ResourceID], output)
	}
	return m, nil
}

func outputsByKind(outputs []types.Output) (map[string][]types.Output, error) {
	m := make(map[string][]types.Output)
	for _, output := range outputs {
		key, err := decodeOutputKey(output)
		if err != nil {
			return nil, fmt.Errorf("failed to decode output key: %w", err)
		}
		m[key.ResourceKind] = append(m[key.ResourceKind], output)
	}
	return m, nil
}

func outputsByPropertyName(outputs []types.Output) (map[string]types.Output, error) {
	m := make(map[string]types.Output)
	for _, output := range outputs {
		key, err := decodeOutputKey(output)
		if err != nil {
			return nil, fmt.Errorf("failed to decode output key: %w", err)
		}
		m[key.PropertyName] = output
	}
	return m, nil
}

func (c *CloudformationProvisioner) updateResources(ctx context.Context, outputs []types.Output) ([]*provisioner.ProvisioningEvent, error) {
	byKind, err := outputsByKind(outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to group outputs by kind: %w", err)
	}

	var events []*provisioner.ProvisioningEvent

	for kind, outputs := range byKind {
		byResourceID, err := outputsByResourceID(outputs)
		if err != nil {
			return nil, fmt.Errorf("failed to group outputs by resource ID: %w", err)
		}
		for id, outputs := range byResourceID {
			switch kind {
			case ResourceKindPostgres:
				e, err := updatePostgresOutputs(ctx, id, outputs)
				if err != nil {
					return nil, fmt.Errorf("failed to update postgres outputs: %w", err)
				}
				events = append(events, e...)
			case ResourceKindMySQL:
				panic("mysql not implemented")
			}
		}
	}
	return events, nil
}

func endpointToDSN(endpoint *string, database string, port int, username, password string) string {
	url := url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", *endpoint, port),
		Path:   database,
	}

	query := url.Query()
	query.Add("user", username)
	query.Add("password", password)
	url.RawQuery = query.Encode()

	return url.String()
}
