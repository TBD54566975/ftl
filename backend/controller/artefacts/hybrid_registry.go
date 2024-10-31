package artefacts

import (
	"context"
	"fmt"
	"io"

	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
)

type hybridRegistry struct {
	container *containerRegistry
	dal       *dalRegistry
	Handle    *libdal.Handle[hybridRegistry]
}

func newHybridRegistry(c ContainerConfig, conn libdal.Connection) *hybridRegistry {
	return &hybridRegistry{
		container: newContainerRegistry(c, conn),
		dal:       newDALRegistry(conn),
		Handle: libdal.New(conn, func(h *libdal.Handle[hybridRegistry]) *hybridRegistry {
			return &hybridRegistry{
				container: newContainerRegistry(c, h.Connection),
				dal:       newDALRegistry(h.Connection),
				Handle:    h,
			}
		}),
	}
}

// GetDigestsKeys locates the `ArtefactKey` for each digest from the container store and database store. The entries
// located on the container store take precedent.
func (s *hybridRegistry) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
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

func (s *hybridRegistry) Upload(ctx context.Context, artefact Artefact) (sha256.SHA256, error) {
	return s.container.Upload(ctx, artefact)
}

func (s *hybridRegistry) Download(ctx context.Context, digest sha256.SHA256) (io.ReadCloser, error) {
	present, _, err := s.container.GetDigestsKeys(ctx, []sha256.SHA256{digest})
	if err != nil {
		return nil, fmt.Errorf("unable to verify artifact's (%s) presence in the container store: %w", digest, err)
	}
	if len(present) == 1 {
		return s.container.Download(ctx, digest)
	}
	return s.dal.Download(ctx, digest)
}

func (s *hybridRegistry) GetReleaseArtefacts(ctx context.Context, releaseID int64) ([]ReleaseArtefact, error) {
	if ras, err := s.dal.GetReleaseArtefacts(ctx, releaseID); err != nil {
		return nil, fmt.Errorf("unable to get release artefacts from container store: %w", err)
	} else if len(ras) > 0 {
		return ras, nil
	}
	return s.dal.GetReleaseArtefacts(ctx, releaseID)
}

func (s *hybridRegistry) AddReleaseArtefact(ctx context.Context, key model.DeploymentKey, ra ReleaseArtefact) error {
	return s.container.AddReleaseArtefact(ctx, key, ra)
}
