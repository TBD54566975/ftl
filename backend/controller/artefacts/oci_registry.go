package artefacts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	googleremote "github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/internal/log"
)

var _ Service = &OCIArtefactService{}

type RegistryConfig struct {
	Registry      string `help:"OCI container registry, in the form host[:port]/repository" env:"FTL_ARTEFACT_REGISTRY" required:""`
	Username      string `help:"OCI container registry username" env:"FTL_ARTEFACT_REGISTRY_USERNAME"`
	Password      string `help:"OCI container registry password" env:"FTL_ARTEFACT_REGISTRY_PASSWORD"`
	AllowInsecure bool   `help:"Allows the use of insecure HTTP based registries." env:"FTL_ARTEFACT_REGISTRY_ALLOW_INSECURE"`
}

type OCIArtefactService struct {
	config RegistryConfig
	auth   authn.AuthConfig
	puller *googleremote.Puller
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

func NewForTesting() *OCIArtefactService {
	storage, err := NewOCIRegistryStorage(RegistryConfig{Registry: "127.0.0.1:15000/ftl-tests", AllowInsecure: true})
	if err != nil {
		panic(err)
	}
	return storage
}

func NewOCIRegistryStorage(config RegistryConfig) (*OCIArtefactService, error) {
	// Connect the registry targeting the specified container
	puller, err := googleremote.NewPuller()
	if err != nil {
		return nil, fmt.Errorf("unable to create puller for registry '%s': %w", config.Registry, err)
	}
	return &OCIArtefactService{
		config: config,
		auth:   authn.AuthConfig{Username: config.Username, Password: config.Password},
		puller: puller,
	}, nil
}

func (s *OCIArtefactService) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	repo, err := s.repoFactory()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to container registry '%s': %w", s.config.Registry, err)
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
func (s *OCIArtefactService) Upload(ctx context.Context, artefact Artefact) (sha256.SHA256, error) {
	repo, err := s.repoFactory()
	logger := log.FromContext(ctx)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to connect to repository '%s': %w", s.config.Registry, err)
	}

	// 2. Pack the files and tag the packed manifest
	artifactType := "application/vnd.ftl.artifact"

	store := memory.New()
	desc, err := pushBlob(ctx, artifactType, artefact.Content, store)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to push to in memory repository %w", err)
	}
	tag := desc.Digest.Hex()
	parseSHA256, err := sha256.ParseSHA256(tag)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to parse sha %w", err)
	}
	artefact.Digest = parseSHA256
	logger.Debugf("Tagging module blob with digest '%s'", tag)

	fileDescriptors := []ocispec.Descriptor{desc}
	var configBlob []byte
	configDesc, err := pushBlob(ctx, ocispec.MediaTypeImageConfig, configBlob, store) // push config blob
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to push config to in memory repository %w", err)
	}
	manifestBlob, err := generateManifestContent(configDesc, fileDescriptors...)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to generate manifest content %w", err)
	}
	manifestDesc, err := pushBlob(ctx, ocispec.MediaTypeImageManifest, manifestBlob, store) // push manifest blob
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to push manifest to in memory repository %w", err)
	}
	if err = store.Tag(ctx, manifestDesc, tag); err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to tag in memory repository %w", err)
	}

	// 4. Copy from the file store to the remote repository
	_, err = oras.Copy(ctx, store, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		return sha256.SHA256{}, fmt.Errorf("unable to upload artifact: %w", err)
	}

	return artefact.Digest, nil
}

func (s *OCIArtefactService) Download(ctx context.Context, dg sha256.SHA256) (io.ReadCloser, error) {
	// ORAS is really annoying, and needs you to know the size of the blob you're downloading
	// So we are using google's go-containerregistry to do the actual download
	// This is not great, we should remove oras at some point
	opts := []name.Option{}
	if s.config.AllowInsecure {
		opts = append(opts, name.Insecure)
	}
	newDigest, err := name.NewDigest(fmt.Sprintf("%s@sha256:%s", s.config.Registry, dg.String()), opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create digest '%s': %w", dg, err)
	}
	layer, err := googleremote.Layer(newDigest, googleremote.WithAuthFromKeychain(authn.DefaultKeychain), googleremote.Reuse(s.puller))
	if err != nil {
		return nil, fmt.Errorf("unable to read layer '%s': %w", newDigest, err)
	}
	uncompressed, err := layer.Uncompressed()
	if err != nil {
		return nil, fmt.Errorf("unable to read uncompressed layer '%s': %w", newDigest, err)
	}
	return uncompressed, nil
}

func (s *OCIArtefactService) repoFactory() (*remote.Repository, error) {
	reg, err := remote.NewRepository(s.config.Registry)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to container registry '%s': %w", s.config.Registry, err)
	}

	reg.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(s.config.Registry, auth.Credential{
			Username: s.config.Username,
			Password: s.config.Password,
		}),
	}
	reg.PlainHTTP = s.config.AllowInsecure
	return reg, nil
}

func pushBlob(ctx context.Context, mediaType string, blob []byte, target oras.Target) (desc ocispec.Descriptor, err error) {
	desc = ocispec.Descriptor{ // Generate descriptor based on the media type and blob content
		MediaType: mediaType,
		Digest:    digest.FromBytes(blob), // Calculate digest
		Size:      int64(len(blob)),       // Include blob size
	}
	err = target.Push(ctx, desc, bytes.NewReader(blob)) // Push the blob to the registry target
	if err != nil {
		return desc, fmt.Errorf("unable to push blob: %w", err)
	}
	return desc, nil
}

func generateManifestContent(config ocispec.Descriptor, layers ...ocispec.Descriptor) ([]byte, error) {
	content := ocispec.Manifest{
		Config:    config, // Set config blob
		Layers:    layers, // Set layer blobs
		Versioned: specs.Versioned{SchemaVersion: 2},
	}
	json, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal manifest content: %w", err)
	}
	return json, nil // Get json content
}
