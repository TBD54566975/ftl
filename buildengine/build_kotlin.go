package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beevik/etree"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

func buildKotlin(ctx context.Context, _ *schema.Schema, module Module) error {
	logger := log.FromContext(ctx)

	if err := SetPOMProperties(ctx, filepath.Join(module.Dir, "..")); err != nil {
		return fmt.Errorf("unable to update ftl.version in %s: %w", module.Dir, err)
	}

	logger.Debugf("Using build command '%s'", module.Build)
	err := exec.Command(ctx, log.Debug, module.Dir, "bash", "-c", module.Build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module: %w", err)
	}

	return nil
}

// SetPOMProperties updates the ftl.version and ftlEndpoint properties in the
// pom.xml file in the given base directory.
func SetPOMProperties(ctx context.Context, baseDir string) error {
	logger := log.FromContext(ctx)
	ftlVersion := ftl.Version
	if ftlVersion == "dev" {
		ftlVersion = "1.0-SNAPSHOT"
	}

	ftlEndpoint := os.Getenv("FTL_ENDPOINT")
	if ftlEndpoint == "" {
		ftlEndpoint = "http://127.0.0.1:8892"
	}

	pomFile := filepath.Clean(filepath.Join(baseDir, "pom.xml"))

	logger.Debugf("Setting ftl.version in %s to %s", pomFile, ftlVersion)

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
		logger.Warnf("unable to find <properties>/<ftlEndpoint> in %s", pomFile)
	} else {
		endpoint.SetText(ftlEndpoint)
	}

	return tree.WriteToFile(pomFile)
}
