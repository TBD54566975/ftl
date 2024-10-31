package artefacts

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

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
	"github.com/TBD54566975/ftl/internal/slices"
)

const (
	ModuleBlobsPrefix = "ftl/modules"
)

type ContainerConfig struct {
	Registry          string `help:"OCI container registry host:port" default:"127.0.0.1:5001"  env:"FTL_ARTEFACTS_REGISTRY"`
	RegistryUsername  string `help:"OCI container registry username" env:"FTL_ARTEFACTS_USER"`
	RegistryPassword  string `help:"OCI container registry password" env:"FTL_ARTEFACTS_PWD"`
	RegistryAllowHTTP bool   `help:"Allows OCI container requests to accept plain HTTP responses" default:"true" env:"FTL_ARTEFACTS_ALLOW_HTTP"`
}

type containerRegistry struct {
	host        string
	repoFactory func() (*remote.Repository, error)

	// in the interim releases and artefacts will continue to be linked via the `deployment_artefacts` table
	Handle *libdal.Handle[containerRegistry]
	db     sql.Querier
}

func newContainerRegistry(c ContainerConfig, conn libdal.Connection) *containerRegistry {
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
				Username: c.RegistryUsername,
				Password: c.RegistryPassword,
			}),
		}
		reg.PlainHTTP = c.RegistryAllowHTTP

		return reg, nil
	}

	return &containerRegistry{
		host:        c.Registry,
		repoFactory: repoFactory,
		Handle: libdal.New(conn, func(h *libdal.Handle[containerRegistry]) *containerRegistry {
			return &containerRegistry{
				host:        c.Registry,
				repoFactory: repoFactory,
				Handle:      h,
				db:          sql.New(h.Connection),
			}
		}),
	}
}

func (s *containerRegistry) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
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
func (s *containerRegistry) Upload(ctx context.Context, artefact Artefact) (sha256.SHA256, error) {
	repo, err := s.repoFactory()
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to connect to repository '%s': %w", s.host, err)
	}
	desc := content.NewDescriptorFromBytes("application/x-octet-stream", artefact.Content)
	if err = repo.Push(ctx, desc, bytes.NewReader(artefact.Content)); err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to upload module blob to repository: %w", err)
	}
	return artefact.Digest, nil
}

func (s *containerRegistry) Download(ctx context.Context, digest sha256.SHA256) (io.ReadCloser, error) {
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

func (s *containerRegistry) GetReleaseArtefacts(ctx context.Context, releaseID int64) ([]ReleaseArtefact, error) {
	rows, err := s.db.GetReleaseArtefacts(ctx, releaseID)
	if err != nil {
		return nil, fmt.Errorf("unable to get release artefacts: %w", libdal.TranslatePGError(err))
	}
	return slices.Map(rows, func(row sql.GetReleaseArtefactsRow) ReleaseArtefact {
		return ReleaseArtefact{
			Artefact:   ArtefactKey{Digest: sha256.FromBytes(row.Digest)},
			Path:       row.Path,
			Executable: row.Executable,
		}
	}), nil
}

func (s *containerRegistry) AddReleaseArtefact(ctx context.Context, key model.DeploymentKey, ra ReleaseArtefact) error {
	params := sql.PublishReleaseArtefactParams{
		Key:        key,
		Digest:     ra.Artefact.Digest[:],
		Executable: ra.Executable,
		Path:       ra.Path,
	}
	if err := s.db.PublishReleaseArtefact(ctx, params); err != nil {
		return libdal.TranslatePGError(err)
	}
	return nil
}

// createModuleRepositoryPathFromDigest creates the path to the repository, relative to the registries root
func createModuleRepositoryPathFromDigest(digest sha256.SHA256) string {
	return fmt.Sprintf("%s/%s:latest", ModuleBlobsPrefix, hex.EncodeToString(digest[:]))
}

// createModuleRepositoryReferenceFromDigest creates the URL used to connect to the repository
func createModuleRepositoryReferenceFromDigest(host string, digest sha256.SHA256) string {
	return fmt.Sprintf("%s/%s", host, createModuleRepositoryPathFromDigest(digest))
}
