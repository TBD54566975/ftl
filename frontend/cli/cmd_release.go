package main

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"

	"github.com/block/ftl/backend/controller/artefacts"
	sh "github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
)

type releaseCmd struct {
	Publish  releasePublishCmd        `cmd:"" help:"Packages the project into a release and publishes it."`
	Exists   releaseExistsCmd         `cmd:"" help:"Indicates whether modules, with the specified digests, have been published."`
	Registry artefacts.RegistryConfig `embed:"" prefix:"oci-"`
}

type releasePublishCmd struct {
}

func (d *releasePublishCmd) Run(release *releaseCmd) error {
	svc, err := createContainerService(release)
	if err != nil {
		return fmt.Errorf("failed to create container service: %w", err)
	}
	content := uuid.New()
	contentBytes := content[:]
	hash, err := svc.Upload(context.Background(), artefacts.Artefact{
		Digest: sha256.Sum256(contentBytes),
		Metadata: artefacts.Metadata{
			Path: fmt.Sprintf("random/%s", content),
		},
		Content: contentBytes,
	})
	if err != nil {
		return fmt.Errorf("failed upload artefact: %w", err)
	}
	fmt.Printf("Artefact published with hash: \033[31m%s\033[0m\n", hash)
	return nil
}

type releaseExistsCmd struct {
	Digests []string `help:"Digest sha256:hex" default:""`
}

func (d *releaseExistsCmd) Run(release *releaseCmd) error {
	svc, err := createContainerService(release)
	if err != nil {
		return fmt.Errorf("failed to create container service: %w", err)
	}
	digests := slices.Map(slices.Unique(d.Digests), sh.MustParseSHA256)
	keys, missing, err := svc.GetDigestsKeys(context.Background(), digests)
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}
	fmt.Printf("\033[31m%d\033[0m FTL module blobs located\n", len(keys))
	for i, key := range keys {
		fmt.Printf("  \u001B[34m%02d\u001B[0m - sha256:\u001B[32m%s\u001B[0m\n", i+1, key.Digest)
	}
	if len(missing) > 0 {
		fmt.Printf("\033[31m%d\033[0m FTL module blobs keys \033[31mnot found\033[0m\n", len(missing))
		for i, key := range missing {
			fmt.Printf("  \u001B[34m%02d\u001B[0m - sha256:\u001B[31m%s\u001B[0m\n", i+1, key)
		}
	}
	return nil
}

func createContainerService(release *releaseCmd) (*artefacts.OCIArtefactService, error) {
	storage, err := artefacts.NewOCIRegistryStorage(release.Registry)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCI registry storage: %w", err)
	}
	return storage, nil
}
