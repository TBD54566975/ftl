package state

import (
	"github.com/TBD54566975/ftl/common/sha256"
)

//protobuf:7
type DeploymentArtefact struct {
	Digest     sha256.SHA256 `protobuf:"1"`
	Path       string        `protobuf:"2"`
	Executable bool          `protobuf:"3"`
}

var _ ControllerEvent = (*DeploymentArtefactCreatedEvent)(nil)

//protobuf:6
type DeploymentArtefactCreatedEvent struct {
	Digest sha256.SHA256 `protobuf:"1"`
}

func (d *DeploymentArtefactCreatedEvent) Handle(view State) (State, error) {
	view.artifacts[d.Digest.String()] = true
	return view, nil
}
