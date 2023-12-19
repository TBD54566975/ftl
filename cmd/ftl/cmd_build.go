package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beevik/etree"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
)

type buildCmd struct {
	ModuleDir string `arg:"" help:"Directory containing ftl.toml" type:"existingdir" default:"."`
}

func (b *buildCmd) Run(ctx context.Context) error {
	// Load the TOML file.
	config, err := moduleconfig.LoadConfig(b.ModuleDir)
	if err != nil {
		return err
	}

	switch config.Language {
	case "kotlin":
		return b.buildKotlin(ctx, config)
	default:
		return fmt.Errorf("unable to build. unknown language %q", config.Language)
	}
}

func (b *buildCmd) buildKotlin(ctx context.Context, config moduleconfig.ModuleConfig) error {
	logger := log.FromContext(ctx)

	logger.Infof("Building kotlin module '%s'", config.Module)

	if err := b.setPomProperties(logger); err != nil {
		return fmt.Errorf("unable to update ftl.version in %s: %w", b.ModuleDir, err)
	}

	logger.Infof("Using build command '%s'", config.Build)
	err := exec.Command(ctx, logger.GetLevel(), b.ModuleDir, "bash", "-c", config.Build).Run()
	if err != nil {
		return err
	}

	return nil
}

func (b *buildCmd) setPomProperties(logger *log.Logger) error {
	ftlVersion := ftl.Version
	if ftlVersion == "dev" {
		ftlVersion = "1.0-SNAPSHOT"
	}

	ftlEndpoint := os.Getenv("FTL_ENDPOINT")
	if ftlEndpoint == "" {
		ftlEndpoint = "http://127.0.0.1:8892"
	}

	pomFile := filepath.Clean(filepath.Join(b.ModuleDir, "..", "pom.xml"))

	logger.Infof("Setting ftl.version in %s to %s", pomFile, ftlVersion)

	tree := etree.NewDocument()
	if err := tree.ReadFromFile(pomFile); err != nil {
		return fmt.Errorf("unable to read %s: %w", pomFile, err)
	}
	root := tree.Root()
	properties := root.SelectElement("properties")
	if properties == nil {
		return fmt.Errorf("unable to find <properties> in %s", pomFile)
	}
	version := properties.SelectElement("ftl.version")
	if version == nil {
		return fmt.Errorf("unable to find <properties>/<ftl.version> in %s", pomFile)
	}
	version.SetText(ftlVersion)

	endpoint := properties.SelectElement("ftlEndpoint")
	if endpoint == nil {
		return fmt.Errorf("unable to find <properties>/<ftlEndpoint> in %s", pomFile)
	}
	endpoint.SetText(ftlEndpoint)

	return tree.WriteToFile(pomFile)
}
