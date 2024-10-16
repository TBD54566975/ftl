package main

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	internalobservability "github.com/TBD54566975/ftl/internal/observability"
)

type releaseCmd struct {
	Registry             string `help:"Registry host:port" default:"127.0.0.1:5001"`
	DSN                  string `help:"DAL DSN." default:"postgres://127.0.0.1:15432/ftl?sslmode=disable&user=postgres&password=secret" env:"FTL_CONTROLLER_DSN"`
	MaxOpenDBConnections int    `help:"Maximum number of database connections." default:"20" env:"FTL_MAX_OPEN_DB_CONNECTIONS"`
	MaxIdleDBConnections int    `help:"Maximum number of idle database connections." default:"20" env:"FTL_MAX_IDLE_DB_CONNECTIONS"`

	Publish releasePublishCmd `cmd:"" help:"Packages the project into a release and publishes it."`
	List    releaseListCmd    `cmd:"" help:"Lists all published releases."`
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

type releaseListCmd struct {
}

func (d *releaseListCmd) Run(release *releaseCmd) error {
	svc, err := createContainerService(release)
	if err != nil {
		return fmt.Errorf("failed to create container service: %w", err)
	}
	modules, err := svc.DiscoverModuleArtefacts(context.Background())
	if err != nil {
		return fmt.Errorf("failed to discover module artefacts: %w", err)
	}
	if len(modules) == 0 {
		fmt.Println("No module artefacts found.")
		return nil
	}

	format := "    Digest        : %s\n    Size          : %-7d\n    Repo Digest   : %s\n    Media Type    : %s\n    Artefact Type : %s\n"
	fmt.Printf("Found %d module artefacts:\n", len(modules))
	for i, m := range modules {
		fmt.Printf("\033[31m  Artefact %d\033[0m\n", i)
		fmt.Printf(format, m.ModuleDigest, m.Size, m.RepositoryDigest, m.MediaType, m.ArtefactType)
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
