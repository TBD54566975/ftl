package artefacts

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/opencontainers/go-digest"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/TBD54566975/ftl/backend/controller/artefacts/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
)

const (
	ModuleBlobsPrefix = "ftl/modules/"
)

type ContainerConfig struct {
	Registry       string `help:"OCI container registry host:port" env:"FTL_ARTEFACTS_REGISTRY"`
	Username       string `help:"OCI container registry username" env:"FTL_ARTEFACTS_USER"`
	Password       string `help:"OCI container registry password" env:"FTL_ARTEFACTS_PWD"`
	AllowPlainHTTP bool   `help:"Allows OCI container requests to accept plain HTTP responses" env:"FTL_ARTEFACTS_ALLOW_HTTP"`
}

type ContainerService struct {
	host        string
	repoFactory func() (*remote.Repository, error)

	// in the interim releases and artefacts will continue to be linked via the `deployment_artefacts` table
	Handle *libdal.Handle[ContainerService]
	db     sql.Querier
}

type ArtefactRepository struct {
	ModuleDigest     sha256.SHA256
	MediaType        string
	ArtefactType     string
	RepositoryDigest digest.Digest
	Size             int64
}

type ArtefactBlobs struct {
	Digest    sha256.SHA256
	MediaType string
	Size      int64
}

func NewContainerService(c ContainerConfig, conn libdal.Connection) *ContainerService {
	// Connect the registry targeting the specified container
	repoFactory := func() (*remote.Repository, error) {
		ref := fmt.Sprintf("%s/%s", c.Registry, ModuleBlobsPrefix)
		reg, err := remote.NewRepository(ref)
		if err != nil {
			return nil, fmt.Errorf("unable to connect to container registry '%s': %w", ref, err)
		}

		reg.Client = &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.NewCache(),
			Credential: auth.StaticCredential(c.Registry, auth.Credential{
				Username: c.Username,
				Password: c.Password,
			}),
		}
		reg.PlainHTTP = c.AllowPlainHTTP

		return reg, nil
	}

	return &ContainerService{
		host:        c.Registry,
		repoFactory: repoFactory,
		Handle: libdal.New(conn, func(h *libdal.Handle[ContainerService]) *ContainerService {
			return &ContainerService{
				host:        c.Registry,
				repoFactory: repoFactory,
				Handle:      h,
				db:          sql.New(h.Connection),
			}
		}),
	}
}

func (s *ContainerService) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	repo, err := s.repoFactory()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to container registry '%s': %w", s.host, err)
	}
	set := make(map[sha256.SHA256]bool)
	for _, d := range digests {
		set[d] = true
	}
	keys = make([]ArtefactKey, 0)
	blobs := repo.Blobs()
	for _, d := range digests {
		_, err := blobs.Resolve(ctx, fmt.Sprintf("sha256:%s", d))
		if err != nil {
			if errors.Is(err, errdef.ErrNotFound) {
				continue
			}
			return nil, nil, fmt.Errorf("unable to resolve digest '%s': %w", d, err)
		}
		keys = append(keys, ArtefactKey{Digest: d})
		delete(set, d)
	}
	missing = make([]sha256.SHA256, 0)
	for d := range set {
		missing = append(missing, d)
	}
	return keys, missing, nil
}

// Upload uploads the specific artifact as a raw blob and links it to a manifest to prevent GC
func (s *ContainerService) Upload(ctx context.Context, artefact Artefact) (sha256.SHA256, error) {
	repo, err := s.repoFactory()
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to connect to repository '%s/%s': %w", s.host, err)
	}
	desc := content.NewDescriptorFromBytes("application/x-octet-stream", artefact.Content)
	if err = repo.Push(ctx, desc, bytes.NewReader(artefact.Content)); err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to upload module blob to repository: %w", err)
	}
	return artefact.Digest, nil
}

func (s *ContainerService) Download(ctx context.Context, digest sha256.SHA256) (io.ReadCloser, error) {
	ref := createModuleRepositoryPathFromDigest(digest)
	registry, err := s.repoFactory()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to registry '%s/%s': %w", s.host, ref, err)
	}
	_, stream, err := oras.Fetch(ctx, registry, createModuleRepositoryReferenceFromDigest(s.host, digest), oras.DefaultFetchOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to download artefact: %w", err)
	}
	return stream, nil
}

func (s *ContainerService) GetReleaseArtefacts(ctx context.Context, releaseID int64) ([]ReleaseArtefact, error) {
	return getDatabaseReleaseArtefacts(ctx, s.db, releaseID)
}

func (s *ContainerService) AddReleaseArtefact(ctx context.Context, key model.DeploymentKey, ra ReleaseArtefact) error {
	return addReleaseArtefacts(ctx, s.db, key, ra)
}

// createModuleRepositoryPathFromDigest creates the path to the repository, relative to the registries root
func createModuleRepositoryPathFromDigest(digest sha256.SHA256) string {
	return fmt.Sprintf("%s/%s:latest", ModuleBlobsPrefix, hex.EncodeToString(digest[:]))
}

// createModuleRepositoryReferenceFromDigest creates the URL used to connect to the repository
func createModuleRepositoryReferenceFromDigest(host string, digest sha256.SHA256) string {
	return fmt.Sprintf("%s/%s", host, createModuleRepositoryPathFromDigest(digest))
}
