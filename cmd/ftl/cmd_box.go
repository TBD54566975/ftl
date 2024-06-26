package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

const boxDockerFile = `FROM {{.BaseImage}}

WORKDIR /root

COPY modules /root

EXPOSE 8891
EXPOSE 8892

ENTRYPOINT ["/root/ftl", "box-run", "/root/modules"]

`

type boxCmd struct {
	BaseImage   string   `help:"Name of the ftl-box Docker image to use as a base." default:"ftl0/ftl-box:${version}"`
	Parallelism int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Image       string   `arg:"" help:"Name of image to build."`
	Dirs        []string `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
}

func (b *boxCmd) Help() string {
	return ``
}

func (b *boxCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient, projConfig projectconfig.Config) error {
	if len(b.Dirs) == 0 {
		b.Dirs = projConfig.AbsModuleDirs()
	}
	if len(b.Dirs) == 0 {
		return errors.New("no directories specified")
	}
	engine, err := buildengine.New(ctx, client, b.Dirs, buildengine.Parallelism(b.Parallelism))
	if err != nil {
		return err
	}
	if err := os.Setenv("GOOS", "linux"); err != nil {
		return fmt.Errorf("failed to set GOOS: %w", err)
	}
	if err := os.Setenv("GOARCH", "amd64"); err != nil {
		return fmt.Errorf("failed to set GOARCH: %w", err)
	}
	if err := engine.Build(ctx); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	workDir, err := os.MkdirTemp("", "ftl-box-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(workDir) //nolint:errcheck
	logger := log.FromContext(ctx)
	logger.Debugf("Copying")
	if err := engine.Each(func(m buildengine.Module) error {
		config := m.Config.Abs()
		destDir := filepath.Join(workDir, "modules", config.Module)

		// Copy deployment artefacts.
		files, err := buildengine.FindFilesToDeploy(config)
		if err != nil {
			return err
		}
		files = append(files, filepath.Join(config.Dir, "ftl.toml"))
		for _, file := range files {
			relFile, err := filepath.Rel(config.Dir, file)
			if err != nil {
				return err
			}
			destFile := filepath.Join(destDir, relFile)
			logger.Debugf(" %s -> %s", file, destFile)
			if err := copy.Copy(file, destFile); err != nil {
				return fmt.Errorf("failed to copy %s to %s: %w", file, destFile, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	baseImage := b.BaseImage
	baseImageParts := strings.Split(baseImage, ":")
	if len(baseImageParts) == 2 {
		version := baseImageParts[1]
		if !ftl.IsRelease(version) {
			version = "latest"
		}
		baseImage = baseImageParts[0] + ":" + version
	}
	dockerFile := strings.ReplaceAll(boxDockerFile, "{{.BaseImage}}", baseImage)
	err = os.WriteFile(filepath.Join(workDir, "Dockerfile"), []byte(dockerFile), 0600)
	if err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}
	logger.Infof("Building image %s", b.Image)
	return exec.Command(ctx, log.Debug, workDir, "docker", "build", "-t", b.Image, "--progress=plain", "--platform=linux/amd64", ".").RunBuffered(ctx)
}
