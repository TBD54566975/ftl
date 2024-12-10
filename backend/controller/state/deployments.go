package state

import (
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/eventstream"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
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

func (r *State) Deployments() map[string]*Deployment {
	return r.deployments
}

func (r *State) ActiveDeployments() map[string]*Deployment {
	return r.activeDeployments
}

var _ eventstream.Event[State] = (*DeploymentCreatedEvent)(nil)
var _ eventstream.Event[State] = (*DeploymentActivatedEvent)(nil)
var _ eventstream.Event[State] = (*DeploymentDeactivatedEvent)(nil)

type DeploymentCreatedEvent struct {
	Key       model.DeploymentKey
	CreatedAt time.Time
	Module    string
	Schema    *schemapb.Module
	Artefacts []*DeploymentArtefact
	Language  string
}

func (r *DeploymentCreatedEvent) Handle(t State) (State, error) {
	if existing := t.deployments[r.Key.String()]; existing != nil {
		return t, nil
	}
	proto, err := schema.ModuleFromProto(r.Schema)
	if err != nil {
		return t, fmt.Errorf("failed to parse schema: %w", err)
	}
	n := Deployment{
		Key:       r.Key,
		CreatedAt: r.CreatedAt,
		Schema:    proto,
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

type DeploymentActivatedEvent struct {
	Key         model.DeploymentKey
	ActivatedAt time.Time
	MinReplicas int
}

func (r DeploymentActivatedEvent) Handle(t State) (State, error) {
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
	Key model.DeploymentKey
}

func (r DeploymentDeactivatedEvent) Handle(t State) (State, error) {
	existing, ok := t.deployments[r.Key.String()]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)

	}
	existing.MinReplicas = 0
	delete(t.activeDeployments, r.Key.String())
	return t, nil
}
