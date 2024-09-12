package dal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

// NotificationPayload is a row from the database.
//
//sumtype:decl
type NotificationPayload interface{ notification() }

// A Notification from the database.
type Notification[T NotificationPayload, Key any, KeyP interface {
	*Key
	encoding.TextUnmarshaler
}] struct {
	Deleted optional.Option[Key] // If present the object was deleted.
	Message optional.Option[T]
}

func (n Notification[T, Key, KeyP]) String() string {
	if key, ok := n.Deleted.Get(); ok {
		return fmt.Sprintf("deleted %v", key)
	}
	return fmt.Sprintf("message %v", n.Message)
}

// DeploymentNotification is a notification from the database when a deployment changes.
type DeploymentNotification = Notification[Deployment, model.DeploymentKey, *model.DeploymentKey]

type deploymentState struct {
	Key         model.DeploymentKey
	schemaHash  []byte
	minReplicas int
}

func deploymentStateFromDeployment(deployment Deployment) (deploymentState, error) {
	hasher := sha256.New()
	data := []byte(deployment.Schema.String())
	if _, err := hasher.Write(data); err != nil {
		return deploymentState{}, fmt.Errorf("failed to hash schema: %w", err)
	}

	return deploymentState{
		schemaHash:  hasher.Sum(nil),
		minReplicas: deployment.MinReplicas,
		Key:         deployment.Key,
	}, nil
}

func (d *DAL) PollDeployments(ctx context.Context) {
	logger := log.FromContext(ctx)
	retry := backoff.Backoff{}

	previousDeployments := make(map[string]deploymentState)

	for {
		delay := time.Millisecond * 500
		currentDeployments := make(map[string]deploymentState)

		deployments, err := d.GetDeploymentsWithMinReplicas(ctx)
		if err != nil {
			if ctx.Err() == context.Canceled {
				logger.Debugf("Polling stopped: %v", ctx.Err())
				return
			}
			logger.Errorf(err, "failed to get deployments when polling")
			time.Sleep(retry.Duration())
			continue
		}

		// Check for new or updated deployments
		for _, deployment := range deployments {
			name := deployment.Schema.Name
			state, err := deploymentStateFromDeployment(deployment)
			if err != nil {
				logger.Errorf(err, "failed to compute deployment state")
				continue
			}

			currentDeployments[name] = state

			previousState, exists := previousDeployments[name]
			if !exists {
				logger.Tracef("New deployment: %s", name)
				d.DeploymentChanges.Publish(DeploymentNotification{
					Message: optional.Some(deployment),
				})
			} else if !bytes.Equal(previousState.schemaHash, state.schemaHash) || previousState.minReplicas != state.minReplicas || !bytes.Equal(previousState.Key.Suffix, state.Key.Suffix) {
				logger.Tracef("Changed deployment: %s", name)
				d.DeploymentChanges.Publish(DeploymentNotification{
					Message: optional.Some(deployment),
				})
			}
		}

		// Check for removed deployments
		for name := range previousDeployments {
			if _, exists := currentDeployments[name]; !exists {
				logger.Tracef("Removed deployment: %s", name)
				d.DeploymentChanges.Publish(DeploymentNotification{
					Deleted: optional.Some(previousDeployments[name].Key),
				})
			}
		}

		previousDeployments = currentDeployments
		retry.Reset()

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}
