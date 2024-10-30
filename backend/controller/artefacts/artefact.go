package artefacts

import (
	"context"
	"github.com/TBD54566975/ftl/backend/libdal"
	"io"

	"github.com/TBD54566975/ftl/internal/model"
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
}

type ReleaseArtefact struct {
	Artefact   ArtefactKey
	Path       string
	Executable bool
}

type Service interface {
	// GetDigestsKeys locates the `digests` corresponding `ArtefactKey`s and identifies the missing ones
	GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error)
	// Upload pushes the specified media, and metadata, to the registry and returns the computed digest
	Upload(context context.Context, artefact Artefact) (sha256.SHA256, error)
	// Download performs a streaming download of the artefact identified by the supplied digest
	Download(context context.Context, digest sha256.SHA256) (io.ReadCloser, error)
	// GetReleaseArtefacts locates the artefacts metadata corresponding with the specified release
	GetReleaseArtefacts(ctx context.Context, releaseID int64) ([]ReleaseArtefact, error)
	// AddReleaseArtefact associates the given `release` with the artefact associated with the given `digest`
	AddReleaseArtefact(ctx context.Context, key model.DeploymentKey, ra ReleaseArtefact) error
}

func New(c ContainerConfig, conn libdal.Connection) Service {
	return newHybridRegistry(c, conn)
}
