package main

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

func (c *CloudformationProvisioner) Plan(ctx context.Context, req *connect.Request[provisioner.PlanRequest]) (*connect.Response[provisioner.PlanResponse], error) {
	res, updated, err := c.createChangeSet(ctx, req.Msg.Provisioning)
	if err != nil {
		return nil, err
	}
	if !updated {
		// no changes to report
		return connect.NewResponse(&provisioner.PlanResponse{}), nil
	}
	desc, err := c.client.DescribeChangeSet(ctx, &cloudformation.DescribeChangeSetInput{
		ChangeSetName: res.Id,
		StackName:     res.StackId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe change-set %s: %w", *res.Id, err)
	}
	_, err = c.client.DeleteChangeSet(ctx, &cloudformation.DeleteChangeSetInput{
		ChangeSetName: res.Id,
		StackName:     res.StackId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete change-set %s: %w", *res.Id, err)
	}

	bytes, err := json.MarshalIndent(desc.Changes, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to create plan JSON for change-set %s: %w", *res.Id, err)
	}

	return connect.NewResponse(&provisioner.PlanResponse{
		Plan: string(bytes),
	}), nil
}
