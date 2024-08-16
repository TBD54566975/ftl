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
		// This is not a critical error, things will probably work fine
		// TBH updating the pom is maybe not the best idea anyway
		logger.Warnf("unable to update ftl.version in %s: %s", module.Config.Dir, err.Error())
	}
	logger.Infof("Using build command '%s'", module.Config.Build)
	command := exec.Command(ctx, log.Debug, module.Config.Dir, "bash", "-c", module.Config.Build)
	command.Env = append(command.Env, "FTL_MODULE_NAME="+module.Config.Module)
	err := command.RunBuffered(ctx)
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

	parent := root.SelectElement("parent")
	versionSet := false
	if parent != nil {
		// You can't use properties in the parent
		// If they are using our parent then we want to update the version
		group := parent.SelectElement("groupId")
		artifact := parent.SelectElement("artifactId")
		if group.Text() == "xyz.block.ftl" && (artifact.Text() == "ftl-build-parent-java" || artifact.Text() == "ftl-build-parent-kotlin") {
			version := parent.SelectElement("version")
			if version != nil {
				version.SetText(ftlVersion)
				versionSet = true
			}
		}
	}

	err := updatePomProperties(root, pomFile, ftlVersion)
	if err != nil && !versionSet {
		// This is only a failure if we also did not update the parent
		return err
	}

	err = tree.WriteToFile(pomFile)
	if err != nil {
		return fmt.Errorf("unable to write %s: %w", pomFile, err)
	}
	return nil
}

func updatePomProperties(root *etree.Element, pomFile string, ftlVersion string) error {
	properties := root.SelectElement("properties")
	if properties == nil {
		return fmt.Errorf("unable to find <properties> in %s", pomFile)
	}
	version := properties.SelectElement("ftl.version")
	if version == nil {
		return fmt.Errorf("unable to find <properties>/<ftl.version> in %s", pomFile)
	}
	version.SetText(ftlVersion)
	return nil
}
