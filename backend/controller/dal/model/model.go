package model

import (
	"encoding"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/dal/internal/sql"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
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

// Controller states.
const (
	ControllerStateLive = ControllerState(sql.ControllerStateLive)
	ControllerStateDead = ControllerState(sql.ControllerStateDead)
)

type RequestOrigin string

const (
	RequestOriginIngress = RequestOrigin(sql.OriginIngress)
	RequestOriginCron    = RequestOrigin(sql.OriginCron)
	RequestOriginPubsub  = RequestOrigin(sql.OriginPubsub)
)

type Deployment struct {
	Key         model.DeploymentKey
	Language    string
	Module      string
	MinReplicas int
	Replicas    optional.Option[int] // Depending on the query this may or may not be populated.
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
	Controllers   []Controller
	Runners       []Runner
	Deployments   []Deployment
	IngressRoutes []IngressRouteEntry
}

type IngressRoute struct {
	Runner     model.RunnerKey
	Deployment model.DeploymentKey
	Endpoint   string
	Path       string
	Module     string
	Verb       string
}

type IngressRouteEntry struct {
	Deployment model.DeploymentKey
	Module     string
	Verb       string
	Method     string
	Path       string
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
