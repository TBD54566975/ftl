package artefacts

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/TBD54566975/ftl/backend/controller/artefacts/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
)

const (
	ModuleArtifactPrefix = "ftl/modules/"
)

type ContainerConfig struct {
	Registry       string
	Username       string
	Password       string
	AllowPlainHTTP bool
}

type ContainerService struct {
	host                  string
	repoConnectionBuilder func(container string) (*remote.Repository, error)

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

func NewContainerService(c ContainerConfig, conn libdal.Connection) *ContainerService {
	// Connect the registry targeting the specified container
	repoConnectionBuilder := func(path string) (*remote.Repository, error) {
		ref := fmt.Sprintf("%s/%s", c.Registry, path)
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
		host:                  c.Registry,
		repoConnectionBuilder: repoConnectionBuilder,
		Handle: libdal.New(conn, func(h *libdal.Handle[ContainerService]) *ContainerService {
			return &ContainerService{
				host:                  c.Registry,
				repoConnectionBuilder: repoConnectionBuilder,
				Handle:                h,
				db:                    sql.New(h.Connection),
			}
		}),
	}
}

func (s *ContainerService) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	set := make(map[sha256.SHA256]bool)
	for _, d := range digests {
		set[d] = true
	}
	modules, err := s.DiscoverModuleArtefacts(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to discover module artefacts: %w", err)
	}
	keys = make([]ArtefactKey, 0)
	for _, m := range modules {
		if set[m.ModuleDigest] {
			keys = append(keys, ArtefactKey{Digest: m.ModuleDigest})
			delete(set, m.ModuleDigest)
		}
	}
	missing = make([]sha256.SHA256, 0)
	for d := range set {
		missing = append(missing, d)
	}
	return keys, missing, nil
}

func (s *ContainerService) Upload(ctx context.Context, artefact Artefact) (sha256.SHA256, error) {
	hash := sha256.Sum(artefact.Content)
	ref := fmt.Sprintf("ftl/modules/%s", hash)
	ms := memory.New()
	mediaDescriptor := v1.Descriptor{
		MediaType: "application/ftl.module.v1",
		Digest:    digest.NewDigestFromBytes(digest.SHA256, hash[:]),
		Size:      int64(len(artefact.Content)),
	}
	err := ms.Push(ctx, mediaDescriptor, bytes.NewReader(artefact.Content))
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to stage artefact in memory: %w", err)
	}
	artifactType := "application/ftl.module.artifact"
	opts := oras.PackManifestOptions{
		Layers: []v1.Descriptor{mediaDescriptor},
	}
	tag := "latest"
	manifestDescriptor, err := oras.PackManifest(ctx, ms, oras.PackManifestVersion1_1, artifactType, opts)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to pack artifact manifest: %w", err)
	}
	if err = ms.Tag(ctx, manifestDescriptor, tag); err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to tag artifact: %w", err)
	}
	repo, err := s.repoConnectionBuilder(ref)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to connect to repository '%s/%s': %w", s.host, ref, err)
	}
	if _, err = oras.Copy(ctx, ms, tag, repo, tag, oras.DefaultCopyOptions); err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to push artefact upstream from staging: %w", err)
	}
	return hash, nil
}

func (s *ContainerService) Download(ctx context.Context, digest sha256.SHA256) (io.ReadCloser, error) {
	ref := createModuleRepositoryPathFromDigest(digest)
	registry, err := s.repoConnectionBuilder(ref)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to registry '%s/%s': %w", s.host, ref, err)
	}
	_, stream, err := oras.Fetch(ctx, registry, createModuleRepositoryReferenceFromDigest(s.host, digest), oras.DefaultFetchOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to download artefact: %w", err)
	}
	return stream, nil
}

func (s *ContainerService) DiscoverModuleArtefacts(ctx context.Context) ([]ArtefactRepository, error) {
	return s.DiscoverArtefacts(ctx, ModuleArtifactPrefix)
}

func (s *ContainerService) DiscoverArtefacts(ctx context.Context, prefix string) ([]ArtefactRepository, error) {
	registry, err := remote.NewRegistry(s.host)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to registry '%s': %w", s.host, err)
	}
	registry.PlainHTTP = true
	result := make([]ArtefactRepository, 0)
	err = registry.Repositories(ctx, "", func(repos []string) error {
		for _, path := range repos {
			if !strings.HasPrefix(path, prefix) {
				continue
			}
			d, err := getDigestFromModuleRepositoryPath(path)
			if err != nil {
				return fmt.Errorf("unable to get digest from repository path '%s': %w", path, err)
			}
			repo, err := registry.Repository(ctx, path)
			if err != nil {
				return fmt.Errorf("unable to connect to repository '%s': %w", path, err)
			}
			desc, err := repo.Resolve(ctx, "latest")
			if err != nil {
				return fmt.Errorf("unable to resolve module metadata '%s': %w", path, err)
			}
			_, data, err := oras.FetchBytes(ctx, repo, desc.Digest.String(), oras.DefaultFetchBytesOptions)
			if err != nil {
				return fmt.Errorf("unable to fetch module metadata '%s': %w", path, err)
			}
			var manifest v1.Manifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				return fmt.Errorf("unable to unmarshal module metadata '%s': %w", path, err)
			}
			result = append(result, ArtefactRepository{
				ModuleDigest:     d,
				MediaType:        manifest.Layers[0].MediaType,
				ArtefactType:     manifest.ArtifactType,
				RepositoryDigest: desc.Digest,
				Size:             desc.Size,
			})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to discover artefacts: %w", err)
	}
	return result, nil
}

func (s *ContainerService) GetReleaseArtefacts(ctx context.Context, releaseID int64) ([]ReleaseArtefact, error) {
	return getDatabaseReleaseArtefacts(ctx, s.db, releaseID)
}

func (s *ContainerService) AddReleaseArtefact(ctx context.Context, key model.DeploymentKey, ra ReleaseArtefact) error {
	return addReleaseArtefacts(ctx, s.db, key, ra)
}

// createModuleRepositoryPathFromDigest creates the path to the repository, relative to the registries root
func createModuleRepositoryPathFromDigest(digest sha256.SHA256) string {
	return fmt.Sprintf("%s/%s:latest", ModuleArtifactPrefix, hex.EncodeToString(digest[:]))
}

// createModuleRepositoryReferenceFromDigest creates the URL used to connect to the repository
func createModuleRepositoryReferenceFromDigest(host string, digest sha256.SHA256) string {
	return fmt.Sprintf("%s/%s", host, createModuleRepositoryPathFromDigest(digest))
}

// getDigestFromModuleRepositoryPath extracts the digest from the module repository path; e.g. /ftl/modules/<digest>:latest
func getDigestFromModuleRepositoryPath(repository string) (sha256.SHA256, error) {
	slash := strings.LastIndex(repository, "/")
	if slash == -1 {
		return sha256.SHA256{}, fmt.Errorf("unable to parse repository '%s'", repository)
	}
	d, err := sha256.ParseSHA256(repository[slash+1:])
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to parse repository digest '%s': %w", repository, err)
	}
	return d, nil
}
