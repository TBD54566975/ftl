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
	Describe releaseDescribeCmd `cmd:"" help:"Describes the specified release."`
	Publish  releasePublishCmd  `cmd:"" help:"Packages the project into a release and publishes it."`
	List     releaseListCmd     `cmd:"" help:"Lists all published releases."`
}

type releaseDescribeCmd struct {
	Digest string `arg:"" help:"Digest of the target release."`
}

func (d *releaseDescribeCmd) Run() error {
	return fmt.Errorf("release describe not implemented")
}

type releasePublishCmd struct {
	DSN                  string `help:"DAL DSN." default:"postgres://127.0.0.1:15432/ftl?sslmode=disable&user=postgres&password=secret" env:"FTL_CONTROLLER_DSN"`
	MaxOpenDBConnections int    `help:"Maximum number of database connections." default:"20" env:"FTL_MAX_OPEN_DB_CONNECTIONS"`
	MaxIdleDBConnections int    `help:"Maximum number of idle database connections." default:"20" env:"FTL_MAX_IDLE_DB_CONNECTIONS"`
}

func (d *releasePublishCmd) Run() error {
	conn, err := internalobservability.OpenDBAndInstrument(d.DSN)
	if err != nil {
		return fmt.Errorf("failed to open DB connection: %w", err)
	}
	conn.SetMaxIdleConns(d.MaxIdleDBConnections)
	conn.SetMaxOpenConns(d.MaxOpenDBConnections)

	svc := artefacts.NewContainerService(artefacts.ContainerConfig{
		Registry:       "localhost:5000",
		AllowPlainHTTP: true,
	}, conn)
	content := uuid.New()
	contentBytes := content[:]
	_, err = svc.Upload(context.Background(), artefacts.Artefact{
		Digest: sha256.Sum256(contentBytes),
		Metadata: artefacts.Metadata{
			Path: fmt.Sprintf("random/%s", content),
		},
		Content: contentBytes,
	})
	if err != nil {
		return fmt.Errorf("failed upload artefact: %w", err)
	}
	return nil
}

type releaseListCmd struct {
}

func (d *releaseListCmd) Run() error {
	return fmt.Errorf("release list not implemented")
}
