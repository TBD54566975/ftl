package model

import (
	"io"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/common/sha256"
	"github.com/TBD54566975/ftl/backend/schema"
)

type Deployment struct {
	Module    string
	Language  string
	Key       DeploymentKey
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
