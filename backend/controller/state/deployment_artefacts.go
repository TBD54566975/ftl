package state

import (
	"github.com/TBD54566975/ftl/internal/sha256"
)

type DeploymentArtefact struct {
	Digest     sha256.SHA256
	Path       string
	Executable bool
}

var _ ControllerEvent = (*DeploymentArtefactCreatedEvent)(nil)

type DeploymentArtefactCreatedEvent struct {
	Digest sha256.SHA256
}

func (d *DeploymentArtefactCreatedEvent) Handle(view State) (State, error) {
	view.artifacts[d.Digest.String()] = true
	return view, nil
}
