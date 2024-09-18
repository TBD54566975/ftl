package main

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/cmd/provisioners/ftl-provisioner-cloudformation/cfutil"
)

func (c *CloudformationProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	client, err := cfutil.CreateClient(ctx)
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
		return success(&stack)
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
		return success(&stack)
	case types.StackStatusUpdateInProgress:
		return running()
	case types.StackStatusUpdateCompleteCleanupInProgress:
		return running()
	case types.StackStatusUpdateComplete:
		return success(&stack)
	case types.StackStatusUpdateFailed:
		return failure(&stack)
	case types.StackStatusUpdateRollbackInProgress:
		return running()
	default:
		return nil, errors.New("unsupported Cloudformation status code: " + string(desc.Stacks[0].StackStatus))
	}
}

func success(stack *types.Stack) (*connect.Response[provisioner.StatusResponse], error) {
	props, err := propertiesFromOutput(stack.Outputs)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&provisioner.StatusResponse{
		Status:     provisioner.StatusResponse_SUCCEEDED,
		Properties: props,
	}), nil
}

func running() (*connect.Response[provisioner.StatusResponse], error) {
	return connect.NewResponse(&provisioner.StatusResponse{Status: provisioner.StatusResponse_RUNNING}), nil
}

func failure(stack *types.Stack) (*connect.Response[provisioner.StatusResponse], error) {
	return connect.NewResponse(&provisioner.StatusResponse{
		Status:       provisioner.StatusResponse_FAILED,
		ErrorMessage: *stack.StackStatusReason,
	}), nil
}

func propertiesFromOutput(outputs []types.Output) ([]*provisioner.ResourceProperty, error) {
	var result []*provisioner.ResourceProperty
	for _, output := range outputs {
		key, err := cfutil.DecodeOutputKey(output)
		if err != nil {
			return nil, fmt.Errorf("failed to decode output key: %w", err)
		}

		result = append(result, &provisioner.ResourceProperty{
			ResourceId: key.ResourceID,
			Key:        key.PropertyName,
			Value:      *output.OutputValue,
		})
	}
	return result, nil
}
