package model

import (
	"io"

	"github.com/alecthomas/errors"
	"github.com/oklog/ulid/v2"

	"github.com/TBD54566975/ftl/common/sha256"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/schema"
)

type Deployment struct {
	Module    string
	Language  string
	Key       ulid.ULID
	Schema    *schema.Module
	Artefacts []*Artefact
}

// Close is a convenience function to close all artefacts.
func (d *Deployment) Close() error {
	errs := make([]error, 0, len(d.Artefacts))
	for _, a := range d.Artefacts {
		errs = append(errs, a.Content.Close())
	}
	return errors.Join(errs...)
}

type Artefact struct {
	Path       string
	Executable bool
	Digest     sha256.SHA256
	// ~Zero-cost on-demand reader.
	Content io.ReadCloser
}

func (a *Artefact) ToProto() *ftlv1.DeploymentArtefact {
	return &ftlv1.DeploymentArtefact{
		Path:       a.Path,
		Executable: a.Executable,
		Digest:     a.Digest.String(),
	}
}
