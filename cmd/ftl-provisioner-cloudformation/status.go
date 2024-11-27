package main

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
)

func (c *CloudformationProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	token := req.Msg.ProvisioningToken
	// if the task is not in the map, it means that the provisioner has crashed since starting the task
	// in that case, we start a new task to query the existing stack
	task, _ := c.running.LoadOrStore(token, &task{stackID: token})

	if task.err.Load() != nil {
		c.running.Delete(token)
		return nil, connect.NewError(connect.CodeUnknown, task.err.Load())
	}

	if task.outputs.Load() != nil {
		c.running.Delete(token)

		resources := req.Msg.DesiredResources
		if err := c.updateResources(ctx, task.outputs.Load(), resources); err != nil {
			return nil, err
		}
		return connect.NewResponse(&provisioner.StatusResponse{
			Status: &provisioner.StatusResponse_Success{
				Success: &provisioner.StatusResponse_ProvisioningSuccess{
					UpdatedResources: resources,
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

func (c *CloudformationProvisioner) updateResources(ctx context.Context, outputs []types.Output, update []*provisioner.Resource) error {
	byResourceID, err := outputsByResourceID(outputs)
	if err != nil {
		return fmt.Errorf("failed to group outputs by resource ID: %w", err)
	}

	for _, resource := range update {
		if postgres, ok := resource.Resource.(*provisioner.Resource_Postgres); ok {
			if postgres.Postgres == nil {
				postgres.Postgres = &provisioner.PostgresResource{}
			}
			if postgres.Postgres.Output == nil {
				postgres.Postgres.Output = &schemapb.DatabaseRuntime{}
			}

			if err := updatePostgresOutputs(ctx, postgres.Postgres.Output, resource.ResourceId, byResourceID[resource.ResourceId]); err != nil {
				return fmt.Errorf("failed to update postgres outputs: %w", err)
			}
		} else if _, ok := resource.Resource.(*provisioner.Resource_Mysql); ok {
			panic("mysql not implemented")
		}
	}
	return nil
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
