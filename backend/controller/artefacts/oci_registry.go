package artefacts

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/TBD54566975/ftl/backend/controller/artefacts/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
)

type ContainerConfig struct {
	Registry   string
	Repository string
	Username   string
	Password   string
}

type ContainerService struct {
	host       string
	repository *remote.Repository
	// in the interim releases and artefacts will continue to be linked via the `deployment_artefacts` table
	Handle *libdal.Handle[ContainerService]
	db     sql.Querier
}

func NewContainerService(c ContainerConfig, conn libdal.Connection) *ContainerService {
	repository, err := remote.NewRepository(fmt.Sprintf("%s/%s", c.Registry, c.Repository))
	if err != nil {
		panic(fmt.Errorf("unable to connect to OCI repository: %w", err))
	}

	client := &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(c.Registry, auth.Credential{
			Username: c.Username,
			Password: c.Password,
		}),
	}

	s := &ContainerService{
		host:       c.Registry,
		repository: repository,
		Handle: libdal.New(conn, func(h *libdal.Handle[ContainerService]) *ContainerService {
			svc := &ContainerService{
				host:       c.Registry,
				repository: repository,
				Handle:     h,
				db:         sql.New(h.Connection),
			}
			svc.repository.Client = client
			return svc
		}),
	}

	return s
}

func (s *ContainerService) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	return nil, nil, nil
}

func (s *ContainerService) Upload(ctx context.Context, artefact Artefact) (sha256.SHA256, error) {
	_, err := oras.PushBytes(ctx, s.repository, remoteModulePath(s.host, artefact.Digest), artefact.Content)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to upload artefact: %w", err)
	}
	return sha256.SHA256{}, nil
}

func (s *ContainerService) Download(ctx context.Context, digest sha256.SHA256) (io.ReadCloser, error) {
	_, stream, err := oras.Fetch(ctx, s.repository, remoteModulePath(s.host, digest), oras.DefaultFetchOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to download artefact: %w", err)
	}
	return stream, nil
}

func (s *ContainerService) GetReleaseArtefacts(ctx context.Context, releaseID int64) ([]ReleaseArtefact, error) {
	return nil, nil
}

func (s *ContainerService) AddReleaseArtefact(ctx context.Context, key model.DeploymentKey, ra ReleaseArtefact) error {
	return nil
}

func remoteModulePath(host string, digest sha256.SHA256) string {
	return fmt.Sprintf("%s/modules/%s:latest", host, hex.EncodeToString(digest[:]))
}
