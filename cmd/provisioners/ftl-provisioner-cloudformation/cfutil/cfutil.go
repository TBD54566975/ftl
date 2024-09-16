package cfutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/akamensky/base58"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/jpillora/backoff"
)

// EnsureStackExists and if not, creates an empty stack with the givent name
//
// Returns, when the stack is ready
func EnsureStackExists(ctx context.Context, client *cloudformation.Client, name string) error {
	_, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: &name,
	})
	var ae smithy.APIError
	// Not Found is returned as ValidationError from the AWS API
	if errors.As(err, &ae) && ae.ErrorCode() == "ValidationError" {
		empty := `
		{
			"Resources": {
				  "NullResource": {
					"Type": "AWS::CloudFormation::WaitConditionHandle"
				}
			}
		}
		`
		if _, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
			StackName:    &name,
			TemplateBody: &empty,
		}); err != nil {
			return fmt.Errorf("failed to create stack %s: %w", name, err)
		}

		if err := waitStackReady(ctx, client, name); err != nil {
			return fmt.Errorf("stack %s did not become ready: %w", name, err)
		}

	} else if err != nil {
		return fmt.Errorf("failed to describe stack %s: %w", name, err)
	}
	return nil
}

// WaitChangeSetReady returns when the given changeset either became ready, or resulted into no changes error.
func WaitChangeSetReady(ctx context.Context, client *cloudformation.Client, changeSet, stack string) (hadChanges bool, err error) {
	retry := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		desc, err := client.DescribeChangeSet(ctx, &cloudformation.DescribeChangeSetInput{
			ChangeSetName: &changeSet,
			StackName:     &stack,
		})
		if err != nil {
			return false, fmt.Errorf("failed to describe change-set: %w", err)
		}
		if desc.Status == types.ChangeSetStatusFailed {
			// Unfortunately, there does not seem to be a better way to do this
			if *desc.StatusReason == "The submitted information didn't contain changes. Submit different information to create a change set." {
				// clean up the changeset if there were no changes
				_, err := client.DeleteChangeSet(ctx, &cloudformation.DeleteChangeSetInput{
					ChangeSetName: &changeSet,
					StackName:     &stack,
				})
				return false, fmt.Errorf("failed to delete change-set: %w", err)
			}
			return false, errors.New(*desc.StatusReason)
		}
		if desc.Status != types.ChangeSetStatusCreatePending && desc.Status != types.ChangeSetStatusCreateInProgress {
			return true, nil
		}
		time.Sleep(retry.Duration())
	}
}

// CreateClient for interacting with Cloudformation
func CreateClient(ctx context.Context) (*cloudformation.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	return cloudformation.New(
		cloudformation.Options{
			Credentials: cfg.Credentials,
			Region:      cfg.Region,
		},
	), nil
}

// CloudformationOutputKey is structured key to be used as an output from a CF stack
type CloudformationOutputKey struct {
	ResourceID   string `json:"r"`
	PropertyName string `json:"p"`
}

// DecodeOutputKey reads the structured CloudformationOutputKey from the given stack output
func DecodeOutputKey(output types.Output) (*CloudformationOutputKey, error) {
	rawKey := *output.OutputKey
	bytes, err := base58.Decode(rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cloudformation output key: %w", err)
	}
	key := CloudformationOutputKey{}
	if err := json.Unmarshal(bytes, &key); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cloudformation output key: %w", err)
	}
	return &key, nil
}

// AddOutput to the given goformation.Outputs
//
// Encodes the given CloudformationOutputKey, and uses the goformation value as the value.
func AddOutput(to goformation.Outputs, value interface{}, key *CloudformationOutputKey) {
	desc := string(outputKeyJSON(key))
	to[base58.Encode(outputKeyJSON(key))] = goformation.Output{
		Value:       value,
		Description: &desc,
	}
}

func outputKeyJSON(key *CloudformationOutputKey) []byte {
	bytes, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	return bytes
}

func waitStackReady(ctx context.Context, client *cloudformation.Client, name string) error {
	retry := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		state, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
			StackName: &name,
		})
		if err != nil {
			return fmt.Errorf("failed to describe stack: %w", err)
		}
		if state.Stacks[0].StackStatus == types.StackStatusCreateFailed {
			return errors.New(*state.Stacks[0].StackStatusReason)
		}
		if state.Stacks[0].StackStatus != types.StackStatusCreateInProgress {
			return nil
		}
		time.Sleep(retry.Duration())
	}
}
