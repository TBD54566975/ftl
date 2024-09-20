package artefacts

import (
	"context"
	"io"

	"github.com/TBD54566975/ftl/internal/sha256"
)

// Metadata container for an artefact's metadata
type Metadata struct {
	Executable bool
	Path       string
}

// Artefact container for an artefact's payload and metadata
type Artefact struct {
	Digest   sha256.SHA256
	Metadata Metadata
	Content  []byte
}

type ArtefactKey struct {
	Digest sha256.SHA256
	// Temporary
	id int64
}

type Registry interface {
	// GetMissingDigests from the set of specified `digests` identifies which ones are not present in the registry
	GetMissingDigests(context context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error)
	// Upload pushes the specified media, and metadata, to the registry and returns the computed digest
	Upload(context context.Context, artefact Artefact) (sha256.SHA256, error)
	// Download performs a streaming download of the artefact identified by the supplied digest
	Download(context context.Context, key ArtefactKey) io.ReadCloser
}
