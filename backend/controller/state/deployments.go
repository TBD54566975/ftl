package state

import (
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/common/slices"
	"github.com/TBD54566975/ftl/internal/model"
)

type Deployment struct {
	Key         model.DeploymentKey
	Module      string
	Schema      *schema.Module
	MinReplicas int
	CreatedAt   time.Time
	ActivatedAt optional.Option[time.Time]
	Artefacts   map[string]*DeploymentArtefact
	Language    string
}

func (r *State) GetDeployment(deployment model.DeploymentKey) (*Deployment, error) {
	d, ok := r.deployments[deployment.String()]
	if !ok {
		return nil, fmt.Errorf("deployment %s not found", deployment)
	}
	return d, nil
}

func (r *State) GetDeployments() map[string]*Deployment {
	return r.deployments
}

func (r *State) GetActiveDeployments() map[string]*Deployment {
	return r.activeDeployments
}

func (r *State) GetActiveDeploymentSchemas() []*schema.Module {
	rows := r.GetActiveDeployments()
	return slices.Map(maps.Values(rows), func(in *Deployment) *schema.Module { return in.Schema })
}

var _ ControllerEvent = (*DeploymentCreatedEvent)(nil)
var _ ControllerEvent = (*DeploymentActivatedEvent)(nil)
var _ ControllerEvent = (*DeploymentDeactivatedEvent)(nil)
var _ ControllerEvent = (*DeploymentSchemaUpdatedEvent)(nil)
var _ ControllerEvent = (*DeploymentReplicasUpdatedEvent)(nil)

type DeploymentCreatedEvent struct {
	Key       model.DeploymentKey
	CreatedAt time.Time
	Module    string
	Schema    *schema.Module
	Artefacts []*DeploymentArtefact
	Language  string
}

func (r *DeploymentCreatedEvent) Handle(t State) (State, error) {
	if existing := t.deployments[r.Key.String()]; existing != nil {
		return t, nil
	}
	n := Deployment{
		Key:       r.Key,
		CreatedAt: r.CreatedAt,
		Schema:    r.Schema,
		Module:    r.Module,
		Artefacts: map[string]*DeploymentArtefact{},
	}
	for _, a := range r.Artefacts {
		n.Artefacts[a.Digest.String()] = &DeploymentArtefact{
			Digest:     a.Digest,
			Path:       a.Path,
			Executable: a.Executable,
		}
	}
	t.deployments[r.Key.String()] = &n
	return t, nil
}

type DeploymentSchemaUpdatedEvent struct {
	Key    model.DeploymentKey
	Schema *schema.Module
}

func (r *DeploymentSchemaUpdatedEvent) Handle(t State) (State, error) {
	existing, ok := t.deployments[r.Key.String()]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)
	}
	existing.Schema = r.Schema
	return t, nil
}

type DeploymentReplicasUpdatedEvent struct {
	Key      model.DeploymentKey
	Replicas int
}

func (r *DeploymentReplicasUpdatedEvent) Handle(t State) (State, error) {
	existing, ok := t.deployments[r.Key.String()]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)
	}
	if existing.Schema.Runtime == nil {
		existing.Schema.Runtime = &schema.ModuleRuntime{}
	}
	if existing.Schema.Runtime.Scaling == nil {
		existing.Schema.Runtime.Scaling = &schema.ModuleRuntimeScaling{}
	}
	existing.Schema.Runtime.Scaling.MinReplicas = int32(r.Replicas)
	existing.MinReplicas = r.Replicas
	return t, nil
}

type DeploymentActivatedEvent struct {
	Key         model.DeploymentKey
	ActivatedAt time.Time
	MinReplicas int
}

func (r *DeploymentActivatedEvent) Handle(t State) (State, error) {
	existing, ok := t.deployments[r.Key.String()]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)

	}
	existing.ActivatedAt = optional.Some(r.ActivatedAt)
	existing.MinReplicas = r.MinReplicas
	t.activeDeployments[r.Key.String()] = existing
	return t, nil
}

type DeploymentDeactivatedEvent struct {
	Key           model.DeploymentKey
	ModuleRemoved bool
}

func (r *DeploymentDeactivatedEvent) Handle(t State) (State, error) {
	existing, ok := t.deployments[r.Key.String()]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)

	}
	existing.MinReplicas = 0
	delete(t.activeDeployments, r.Key.String())
	return t, nil
}
