package main

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
)

func (c *CloudformationProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	client, err := createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudformation client: %w", err)
	}

	desc, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: &req.Msg.ProvisioningToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe stack: %w", err)
	}
	stack := desc.Stacks[0]

	switch stack.StackStatus {
	case types.StackStatusCreateInProgress:
		return running()
	case types.StackStatusCreateFailed:
		return failure(&stack)
	case types.StackStatusCreateComplete:
		return success(&stack, req.Msg.DesiredResources)
	case types.StackStatusRollbackInProgress:
		return failure(&stack)
	case types.StackStatusRollbackFailed:
		return failure(&stack)
	case types.StackStatusRollbackComplete:
		return failure(&stack)
	case types.StackStatusDeleteInProgress:
		return running()
	case types.StackStatusDeleteFailed:
		return failure(&stack)
	case types.StackStatusDeleteComplete:
		return success(&stack, req.Msg.DesiredResources)
	case types.StackStatusUpdateInProgress:
		return running()
	case types.StackStatusUpdateCompleteCleanupInProgress:
		return running()
	case types.StackStatusUpdateComplete:
		return success(&stack, req.Msg.DesiredResources)
	case types.StackStatusUpdateFailed:
		return failure(&stack)
	case types.StackStatusUpdateRollbackInProgress:
		return running()
	default:
		return nil, errors.New("unsupported Cloudformation status code: " + string(desc.Stacks[0].StackStatus))
	}
}

func success(stack *types.Stack, resources []*provisioner.Resource) (*connect.Response[provisioner.StatusResponse], error) {
	err := updateResources(stack.Outputs, resources)
	if err != nil {
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

func running() (*connect.Response[provisioner.StatusResponse], error) {
	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.StatusResponse_Running{
			Running: &provisioner.StatusResponse_ProvisioningRunning{},
		},
	}), nil
}

func failure(stack *types.Stack) (*connect.Response[provisioner.StatusResponse], error) {
	return nil, connect.NewError(connect.CodeUnknown, errors.New(*stack.StackStatusReason))
}

func updateResources(outputs []types.Output, update []*provisioner.Resource) error {
	for _, output := range outputs {
		key, err := decodeOutputKey(output)
		if err != nil {
			return fmt.Errorf("failed to decode output key: %w", err)
		}
		for _, resource := range update {
			if resource.ResourceId == key.ResourceID {
				if postgres, ok := resource.Resource.(*provisioner.Resource_Postgres); ok {
					if postgres.Postgres == nil {
						postgres.Postgres = &provisioner.PostgresResource{}
					}
					if postgres.Postgres.Output == nil {
						postgres.Postgres.Output = &provisioner.PostgresResource_PostgresResourceOutput{}
					}

					switch key.PropertyName {
					case PropertyDBReadEndpoint:
						postgres.Postgres.Output.ReadDsn = endpointToDSN(*output.OutputValue, key.ResourceID, 5432)
					case PropertyDBWriteEndpoint:
						postgres.Postgres.Output.WriteDsn = endpointToDSN(*output.OutputValue, key.ResourceID, 5432)
					}
				} else if mysql, ok := resource.Resource.(*provisioner.Resource_Mysql); ok {
					if mysql.Mysql == nil {
						mysql.Mysql = &provisioner.MysqlResource{}
					}
					if mysql.Mysql.Output == nil {
						mysql.Mysql.Output = &provisioner.MysqlResource_MysqlResourceOutput{}
					}

					switch key.PropertyName {
					case PropertyDBReadEndpoint:
						mysql.Mysql.Output.ReadDsn = endpointToDSN(*output.OutputValue, key.ResourceID, 5432)
					case PropertyDBWriteEndpoint:
						mysql.Mysql.Output.WriteDsn = endpointToDSN(*output.OutputValue, key.ResourceID, 3306)
					}
				}
			}
		}
	}
	return nil
}

func endpointToDSN(endpoint, database string, port int) string {
	return fmt.Sprintf("postgres://%s:%d/%s?user=postgres&password=password", endpoint, port, database)
}
