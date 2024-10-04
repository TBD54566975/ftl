package artefacts

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/TBD54566975/ftl/backend/controller/artefacts/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type ContainerConfig struct {
	Registry       string
	Username       string
	Password       string
	AllowPlainHTTP bool
}

type ContainerService struct {
	host              string
	connectionBuilder func(container string) (*remote.Repository, error)

	// in the interim releases and artefacts will continue to be linked via the `deployment_artefacts` table
	Handle *libdal.Handle[ContainerService]
	db     sql.Querier
}

func NewContainerService(c ContainerConfig, conn libdal.Connection) *ContainerService {
	// Connect the registry targeting the specified container
	connectionBuilder := func(container string) (*remote.Repository, error) {
		ref := fmt.Sprintf("%s/%s", c.Registry, container)
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
		host:              c.Registry,
		connectionBuilder: connectionBuilder,
		Handle: libdal.New(conn, func(h *libdal.Handle[ContainerService]) *ContainerService {
			return &ContainerService{
				host:              c.Registry,
				connectionBuilder: connectionBuilder,
				Handle:            h,
				db:                sql.New(h.Connection),
			}
		}),
	}
}

func (s *ContainerService) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	return nil, nil, nil
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
	registry, err := s.connectionBuilder(ref)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to connect to registry '%s/%s': %w", s.host, ref, err)
	}
	desc, err := oras.Copy(ctx, ms, tag, registry, tag, oras.DefaultCopyOptions)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to push artefact upstream from staging: %w", err)
	}
	fmt.Printf("hash:%s\ndesc: %s\n%v", hex.EncodeToString(hash[:]), desc.Digest.String(), desc)
	return hash, nil
}

func (s *ContainerService) Download(ctx context.Context, digest sha256.SHA256) (io.ReadCloser, error) {
	ref := fmt.Sprintf("ftl/modules/%s", digest)
	registry, err := s.connectionBuilder(ref)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to registry '%s/%s': %w", s.host, ref, err)
	}
	_, stream, err := oras.Fetch(ctx, registry, remoteModulePath(s.host, digest), oras.DefaultFetchOptions)
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
