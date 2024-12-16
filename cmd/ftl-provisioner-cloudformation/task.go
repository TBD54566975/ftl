package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/jpillora/backoff"

	"github.com/block/ftl/internal/log"
)

type task struct {
	stackID string

	err     atomic.Value[error]
	outputs atomic.Value[[]types.Output]
}

func (t *task) waitForStackReady(ctx context.Context, client *cloudformation.Client) ([]types.Output, error) {
	retry := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		desc, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
			StackName: &t.stackID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe stack: %w", err)
		}
		stack := desc.Stacks[0]

		switch stack.StackStatus {
		// noop while running
		case types.StackStatusCreateInProgress:
		case types.StackStatusUpdateInProgress:
		case types.StackStatusUpdateCompleteCleanupInProgress:
		case types.StackStatusUpdateRollbackInProgress:

		// success
		case types.StackStatusCreateComplete:
			return stack.Outputs, nil
		case types.StackStatusDeleteComplete:
			return stack.Outputs, nil
		case types.StackStatusUpdateComplete:
			return stack.Outputs, nil

		// failures
		case types.StackStatusCreateFailed:
			return nil, fmt.Errorf("stack creation failed: %s", *stack.StackStatusReason)
		case types.StackStatusRollbackInProgress:
			return nil, fmt.Errorf("stack rollback in progress: %s", *stack.StackStatusReason)
		case types.StackStatusRollbackFailed:
			return nil, fmt.Errorf("stack rollback failed: %s", *stack.StackStatusReason)
		case types.StackStatusRollbackComplete:
			return nil, fmt.Errorf("stack rollback complete: %s", *stack.StackStatusReason)
		case types.StackStatusDeleteInProgress:
		case types.StackStatusDeleteFailed:
			return nil, fmt.Errorf("stack deletion failed: %s", *stack.StackStatusReason)
		case types.StackStatusUpdateFailed:
			return nil, fmt.Errorf("stack update failed: %s", *stack.StackStatusReason)
		default:
			return nil, fmt.Errorf("unsupported Cloudformation status code: %s", string(stack.StackStatus))
		}

		time.Sleep(retry.Duration())
	}
}

func (t *task) postUpdate(ctx context.Context, secrets *secretsmanager.Client, outputs []types.Output) error {
	byResourceID, err := outputsByResourceID(outputs)
	if err != nil {
		return fmt.Errorf("failed to group outputs by resource ID: %w", err)
	}

	for resourceID, outputs := range byResourceID {
		byName, err := outputsByPropertyName(outputs)
		if err != nil {
			return fmt.Errorf("failed to group outputs by property name: %w", err)
		}
		if err := PostgresPostUpdate(ctx, secrets, byName, resourceID); err != nil {
			return fmt.Errorf("failed to post-update postgres: %w", err)
		}
	}

	return nil
}

func (t *task) Start(oldCtx context.Context, client *cloudformation.Client, secrets *secretsmanager.Client, changeSetID string) {
	ctx := context.WithoutCancel(oldCtx)
	logger := log.FromContext(ctx)
	go func() {
		if changeSetID != "" {
			logger.Debugf("Executing change-set: %s", changeSetID)
			_, err := client.ExecuteChangeSet(ctx, &cloudformation.ExecuteChangeSetInput{
				ChangeSetName: &changeSetID,
				StackName:     &t.stackID,
			})
			if err != nil {
				logger.Errorf(err, "failed to execute change-set")
				t.err.Store(fmt.Errorf("failed to execute change-set: %w", err))
				return
			}
		}
		logger.Debugf("Waiting for stack to be ready: %s", t.stackID)
		outputs, err := t.waitForStackReady(ctx, client)
		if err != nil {
			logger.Errorf(err, "failed to wait for stack to be ready")
			t.err.Store(err)
			return
		}
		logger.Debugf("Stack ready: %s", t.stackID)
		if err := t.postUpdate(ctx, secrets, outputs); err != nil {
			logger.Errorf(err, "failed to post-update")
			t.err.Store(err)
			return
		}
		logger.Debugf("Post-update complete: %s", t.stackID)
		t.outputs.Store(outputs)
	}()
}

func secretARNToUsernamePassword(ctx context.Context, secrets *secretsmanager.Client, secretARN string) (string, string, error) {
	secret, err := secrets.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretARN,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret value: %w", err)
	}
	secretString := *secret.SecretString

	var secretData map[string]string
	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal secret data: %w", err)
	}

	return secretData["username"], secretData["password"], nil
}
