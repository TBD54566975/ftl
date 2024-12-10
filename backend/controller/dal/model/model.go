package model

import (
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/sha256"
)

type Runner struct {
	Key                model.RunnerKey
	Endpoint           string
	ReservationTimeout optional.Option[time.Duration]
	Module             optional.Option[string]
	Deployment         model.DeploymentKey
	Labels             model.Labels
}

func (Runner) notification() {}

type Reconciliation struct {
	Deployment model.DeploymentKey
	Module     string
	Language   string

	AssignedReplicas int
	RequiredReplicas int
}

type ControllerState string

type RequestOrigin string

type Deployment struct {
	Key         model.DeploymentKey
	Language    string
	Module      string
	MinReplicas int
	Schema      *schema.Module
	CreatedAt   time.Time
	Labels      model.Labels
}

func (d Deployment) String() string { return d.Key.String() }

func (d Deployment) notification() {}

type Controller struct {
	Key      model.ControllerKey
	Endpoint string
	State    ControllerState
}

type Status struct {
	Controllers []Controller
	Runners     []Runner
	Deployments []Deployment
}

type DeploymentArtefact struct {
	Digest     sha256.SHA256
	Executable bool
	Path       string
}

func (d *DeploymentArtefact) ToProto() *ftlv1.DeploymentArtefact {
	return &ftlv1.DeploymentArtefact{
		Digest:     d.Digest.String(),
		Executable: d.Executable,
		Path:       d.Path,
	}
}

func DeploymentArtefactFromProto(in *ftlv1.DeploymentArtefact) (DeploymentArtefact, error) {
	digest, err := sha256.ParseSHA256(in.Digest)
	if err != nil {
		return DeploymentArtefact{}, fmt.Errorf("invalid digest: %w", err)
	}
	return DeploymentArtefact{
		Digest:     digest,
		Executable: in.Executable,
		Path:       in.Path,
	}, nil
}
