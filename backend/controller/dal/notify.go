package dal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"

	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
)

// DeploymentNotification is a notification from the database when a deployment changes.
type DeploymentNotification = dalmodel.Notification[dalmodel.Deployment, model.DeploymentKey, *model.DeploymentKey]

type deploymentState struct {
	Key         model.DeploymentKey
	schemaHash  sha256.SHA256
	minReplicas int
}

func deploymentStateFromDeployment(deployment dalmodel.Deployment) (deploymentState, error) {
	h, err := deployment.Schema.Hash()
	if err != nil {
		return deploymentState{}, fmt.Errorf("failed to hash schema: %w", err)
	}
	return deploymentState{
		schemaHash:  h,
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

		deployments, err := d.GetActiveDeployments()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				logger.Tracef("Polling stopped: %v", ctx.Err())
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
			deploymentName := deployment.Key.String()
			currentDeployments[deploymentName] = state

			previousState, exists := previousDeployments[deploymentName]
			if !exists {
				logger.Tracef("New deployment: %s", name)
				d.DeploymentChanges.Publish(DeploymentNotification{
					Message: optional.Some(deployment),
				})
			} else if previousState.schemaHash != state.schemaHash {
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
