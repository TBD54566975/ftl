package buildengine

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/beevik/etree"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

func buildJavaModule(ctx context.Context, module Module) error {
	logger := log.FromContext(ctx)
	if err := SetPOMProperties(ctx, module.Config.Dir); err != nil {
		return fmt.Errorf("unable to update ftl.version in %s: %w", module.Config.Dir, err)
	}
	logger.Infof("Using build command '%s'", module.Config.Build)
	err := exec.Command(ctx, log.Debug, module.Config.Dir, "bash", "-c", module.Config.Build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", module.Config.Module, err)
	}

	return nil
}

// SetPOMProperties updates the ftl.version properties in the
// pom.xml file in the given base directory.
func SetPOMProperties(ctx context.Context, baseDir string) error {
	logger := log.FromContext(ctx)
	ftlVersion := ftl.Version
	if ftlVersion == "dev" {
		ftlVersion = "1.0-SNAPSHOT"
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

	err := tree.WriteToFile(pomFile)
	if err != nil {
		return fmt.Errorf("unable to write %s: %w", pomFile, err)
	}
	return nil
}
