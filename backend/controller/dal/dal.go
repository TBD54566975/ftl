// Package dal provides a data abstraction layer for the Controller
package dal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	inprocesspubsub "github.com/alecthomas/types/pubsub"
	xmaps "golang.org/x/exp/maps"

	aregistry "github.com/TBD54566975/ftl/backend/controller/artefacts"
	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/pubsub"
	"github.com/TBD54566975/ftl/backend/controller/state"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
)

func New(registry aregistry.Service, state state.ControllerState) *DAL {
	var d *DAL
	d = &DAL{
		registry:          registry,
		state:             state,
		DeploymentChanges: inprocesspubsub.New[DeploymentNotification](),
	}

	return d
}

type DAL struct {
	pubsub   *pubsub.Service
	registry aregistry.Service
	state    state.ControllerState

	// DeploymentChanges is a Topic that receives changes to the deployments table.
	DeploymentChanges *inprocesspubsub.Topic[DeploymentNotification]
}

func (d *DAL) GetActiveDeployments() ([]dalmodel.Deployment, error) {
	view := d.state.View()

	deployments := view.ActiveDeployments()
	return slices.Map(xmaps.Values(deployments), func(in *state.Deployment) dalmodel.Deployment {
		return dalmodel.Deployment{
			Key:         in.Key,
			Module:      in.Module,
			Language:    in.Language,
			MinReplicas: in.MinReplicas,
			Schema:      in.Schema,
		}
	}), nil
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
		err = d.state.Publish(ctx, &state.DeploymentDeactivatedEvent{Key: key})
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
	for _, dep := range view.ActiveDeployments() {
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
	deployments, err := d.GetActiveDeployments()
	if err != nil {
		return nil, err
	}

	schemaMap := map[string]*schema.Module{}
	for _, dep := range deployments {
		if _, ok := schemaMap[dep.Module]; !ok {
			// We only take the older ones
			// If new ones exist they are not live yet
			// Or the old ones would be gone
			schemaMap[dep.Module] = dep.Schema
		}
	}
	fullSchema := &schema.Schema{Modules: xmaps.Values(schemaMap)}
	sch, err := schema.ValidateSchema(fullSchema)
	if err != nil {
		return nil, fmt.Errorf("could not validate schema: %w", err)
	}
	return sch, nil
}

// UpdateModuleSchema updates the schema for a deployment in place.
//
// Note that this is racey as the deployment can be updated by another process. This will go away once we ditch the DB.
func (d *DAL) UpdateModuleSchema(ctx context.Context, deployment model.DeploymentKey, module *schema.Module) error {
	err := d.state.Publish(ctx, &state.DeploymentSchemaUpdatedEvent{
		Key:    deployment,
		Schema: module,
	})
	if err != nil {
		return fmt.Errorf("failed to update deployment schema: %w", err)
	}
	return nil
}

func (d *DAL) GetActiveDeploymentSchemas(ctx context.Context) ([]*schema.Module, error) {
	view := d.state.View()
	rows := view.ActiveDeployments()
	return slices.Map(xmaps.Values(rows), func(in *state.Deployment) *schema.Module { return in.Schema }), nil
}

// GetActiveDeploymentSchemasByDeploymentKey returns the schema for all active deployments by deployment key.
//
// model.DeploymentKey is not used directly as a key as it's not a valid map key.
func (d *DAL) GetActiveDeploymentSchemasByDeploymentKey(ctx context.Context) (map[string]*schema.Module, error) {
	view := d.state.View()
	rows := view.ActiveDeployments()
	return maps.MapValues[string, *state.Deployment, *schema.Module](rows, func(dep string, in *state.Deployment) *schema.Module {
		return in.Schema
	}), nil
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

func sha256esToBytes(digests []sha256.SHA256) [][]byte {
	return slices.Map(digests, func(digest sha256.SHA256) []byte { return digest[:] })
}
