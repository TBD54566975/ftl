package artefacts

import (
	"context"
	"fmt"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
	"io"
)

type HybridService struct {
	container *ContainerService
	dal       *Service
	Handle    *libdal.Handle[HybridService]
}

func NewHybridService(c ContainerConfig, conn libdal.Connection) *HybridService {
	return &HybridService{
		container: NewContainerService(c, conn),
		dal:       New(conn),
		// TODO: discover a way to create an object graph from the TXN connection
		Handle: nil,
	}
}

// GetDigestsKeys locates the `ArtefactKey` for each digest from the container store and database store. The entries
// located on the container store take precedent.
func (s *HybridService) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	ck, cm, err := s.container.GetDigestsKeys(ctx, digests)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get digests keys from container store: %w", err)
	}
	dk, dm, err := s.dal.GetDigestsKeys(ctx, cm)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get digests keys from database store: %w", err)
	}
	return append(ck, dk...), dm, nil
}

func (s *HybridService) Upload(ctx context.Context, artefact Artefact) (sha256.SHA256, error) {
	return s.container.Upload(ctx, artefact)
}

func (s *HybridService) Download(ctx context.Context, digest sha256.SHA256) (io.ReadCloser, error) {
	present, _, err := s.container.GetDigestsKeys(ctx, []sha256.SHA256{digest})
	if err != nil {
		return nil, fmt.Errorf("unable to verify artifact's (%s) presence in the container store: %w", digest, err)
	}
	if len(present) == 1 {
		return s.container.Download(ctx, digest)
	}
	return s.dal.Download(ctx, digest)
}

func (s *HybridService) GetReleaseArtefacts(ctx context.Context, releaseID int64) ([]ReleaseArtefact, error) {
	// note: the container and database store currently use release_artefacts to associated
	return s.container.GetReleaseArtefacts(ctx, releaseID)
}

func (s *HybridService) AddReleaseArtefact(ctx context.Context, key model.DeploymentKey, ra ReleaseArtefact) error {
	return s.container.AddReleaseArtefact(ctx, key, ra)
}
