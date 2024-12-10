// Package dal provides a data abstraction layer for the Controller
package dal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	xmaps "golang.org/x/exp/maps"

	aregistry "github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/controller/pubsub"
	"github.com/TBD54566975/ftl/backend/controller/state"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

func New(registry aregistry.Service, state state.ControllerState) *DAL {
	var d *DAL
	d = &DAL{
		registry: registry,
		state:    state,
	}

	return d
}

type DAL struct {
	pubsub   *pubsub.Service
	registry aregistry.Service
	state    state.ControllerState
}

// SetDeploymentReplicas activates the given deployment.
func (d *DAL) SetDeploymentReplicas(ctx context.Context, key model.DeploymentKey, minReplicas int) (err error) {

	view := d.state.View()
	deployment, err := view.GetDeployment(key)
	if err != nil {
		return fmt.Errorf("could not get deployment: %w", err)
	}

	err = d.state.Publish(ctx, &state.DeploymentReplicasUpdatedEvent{Key: key, Replicas: minReplicas})
	if err != nil {
		return libdal.TranslatePGError(err)
	}
	if minReplicas == 0 {
		err = d.state.Publish(ctx, &state.DeploymentDeactivatedEvent{Key: key, ModuleRemoved: true})
		if err != nil {
			return libdal.TranslatePGError(err)
		}
	} else if deployment.MinReplicas == 0 {
		err = d.state.Publish(ctx, &state.DeploymentActivatedEvent{Key: key, ActivatedAt: time.Now(), MinReplicas: minReplicas})
		if err != nil {
			return libdal.TranslatePGError(err)
		}
	}
	timeline.ClientFromContext(ctx).Publish(ctx, timeline.DeploymentUpdated{
		DeploymentKey:   key,
		MinReplicas:     minReplicas,
		PrevMinReplicas: int(deployment.MinReplicas),
	})

	return nil
}

var ErrReplaceDeploymentAlreadyActive = errors.New("deployment already active")

// ReplaceDeployment replaces an old deployment of a module with a new deployment.
//
// returns ErrReplaceDeploymentAlreadyActive if the new deployment is already active.
func (d *DAL) ReplaceDeployment(ctx context.Context, newDeploymentKey model.DeploymentKey, minReplicas int) (err error) {
	view := d.state.View()
	newDeployment, err := view.GetDeployment(newDeploymentKey)
	if err != nil {
		return fmt.Errorf("replace deployment failed to get deployment for %v: %w", newDeploymentKey, libdal.TranslatePGError(err))
	}

	err = d.state.Publish(ctx, &state.DeploymentActivatedEvent{Key: newDeploymentKey, ActivatedAt: time.Now(), MinReplicas: minReplicas})
	if err != nil {
		return libdal.TranslatePGError(err)
	}

	// If there's an existing deployment, set its desired replicas to 0
	var replacedDeploymentKey optional.Option[model.DeploymentKey]
	// TODO: remove all this, it needs to be event driven
	var oldDeployment *state.Deployment
	for _, dep := range view.GetActiveDeployments() {
		if dep.Module == newDeployment.Module {
			oldDeployment = dep
			break
		}
	}
	if oldDeployment != nil {
		if oldDeployment.Key.String() == newDeploymentKey.String() {
			return fmt.Errorf("replace deployment failed: deployment already exists from %v to %v: %w", oldDeployment.Key, newDeploymentKey, ErrReplaceDeploymentAlreadyActive)
		}
		err = d.state.Publish(ctx, &state.DeploymentReplicasUpdatedEvent{Key: newDeploymentKey, Replicas: minReplicas})
		if err != nil {
			return fmt.Errorf("replace deployment failed to set new deployment replicas from %v to %v: %w", oldDeployment.Key, newDeploymentKey, libdal.TranslatePGError(err))
		}
		err = d.state.Publish(ctx, &state.DeploymentDeactivatedEvent{Key: oldDeployment.Key})
		if err != nil {
			return libdal.TranslatePGError(err)
		}
		replacedDeploymentKey = optional.Some(oldDeployment.Key)
	} else {
		// Set the desired replicas for the new deployment
		err = d.state.Publish(ctx, &state.DeploymentReplicasUpdatedEvent{Key: newDeploymentKey, Replicas: minReplicas})
		if err != nil {
			return fmt.Errorf("replace deployment failed to set replicas for %v: %w", newDeploymentKey, libdal.TranslatePGError(err))
		}
	}

	timeline.ClientFromContext(ctx).Publish(ctx, timeline.DeploymentCreated{
		DeploymentKey:      newDeploymentKey,
		Language:           newDeployment.Language,
		ModuleName:         newDeployment.Module,
		MinReplicas:        minReplicas,
		ReplacedDeployment: replacedDeploymentKey,
	})
	if err != nil {
		return fmt.Errorf("replace deployment failed to create event: %w", libdal.TranslatePGError(err))
	}
	return nil
}

// GetActiveSchema returns the schema for all active deployments.
func (d *DAL) GetActiveSchema(ctx context.Context) (*schema.Schema, error) {
	view := d.state.View()
	deployments := view.GetActiveDeployments()

	schemaMap := map[string]*schema.Module{}
	timeMap := map[string]time.Time{}
	for _, dep := range deployments {
		if _, ok := schemaMap[dep.Module]; ok {
			if timeMap[dep.Module].Before(dep.CreatedAt) {
				continue
			}
		}
		// We only take the older ones
		// If new ones exist they are not live yet
		// Or the old ones would be gone
		schemaMap[dep.Module] = dep.Schema
		timeMap[dep.Module] = dep.CreatedAt
	}
	fullSchema := &schema.Schema{Modules: xmaps.Values(schemaMap)}
	sch, err := schema.ValidateSchema(fullSchema)
	if err != nil {
		return nil, fmt.Errorf("could not validate schema: %w", err)
	}
	return sch, nil
}

type ProcessRunner struct {
	Key      model.RunnerKey
	Endpoint string
	Labels   model.Labels
}

type Process struct {
	Deployment  model.DeploymentKey
	MinReplicas int
	Labels      model.Labels
	Runner      optional.Option[ProcessRunner]
}
