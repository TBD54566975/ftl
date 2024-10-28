package main

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	internalobservability "github.com/TBD54566975/ftl/internal/observability"
	sh "github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
)

type releaseCmd struct {
	Registry             string `help:"Registry host:port" default:"127.0.0.1:5001"`
	DSN                  string `help:"DAL DSN." default:"postgres://127.0.0.1:15432/ftl?sslmode=disable&user=postgres&password=secret" env:"FTL_CONTROLLER_DSN"`
	MaxOpenDBConnections int    `help:"Maximum number of database connections." default:"20" env:"FTL_MAX_OPEN_DB_CONNECTIONS"`
	MaxIdleDBConnections int    `help:"Maximum number of idle database connections." default:"20" env:"FTL_MAX_IDLE_DB_CONNECTIONS"`

	Publish releasePublishCmd `cmd:"" help:"Packages the project into a release and publishes it."`
	Exists  releaseExistsCmd  `cmd:"" help:"Indicates whether modules, with the specified digests, have been published."`
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

func createContainerService(release *releaseCmd) (*artefacts.ContainerService, error) {
	conn, err := internalobservability.OpenDBAndInstrument(release.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}
	conn.SetMaxIdleConns(release.MaxIdleDBConnections)
	conn.SetMaxOpenConns(release.MaxOpenDBConnections)

	return artefacts.NewContainerService(artefacts.ContainerConfig{
		Registry:       release.Registry,
		AllowPlainHTTP: true,
	}, conn), nil
}
