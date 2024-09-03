package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/otiai10/copy"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

// Test locally by running:
//
// Rebuild the image:
//	docker build -t ftl0/ftl-box:latest --platform=linux/amd64 -f Dockerfile.box .
//
// Build the box:
//	ftl box echo --compose=echo-compose.yml

const boxftlProjectFile = `module-dirs = ["/root/modules"]
`

const boxDockerFile = `FROM {{.BaseImage}}

WORKDIR /root

COPY modules /root/modules
COPY ftl-project.toml /root

EXPOSE 8891
EXPOSE 8892

ENTRYPOINT ["/root/ftl", "box-run", "/root/modules"]

`

const boxComposeFile = `name: {{.Name}}-box
services:
  db:
    image: postgres
    platform: linux/{{.GOARCH}}
    command: postgres
    user: postgres
    restart: always
    environment:
      POSTGRES_PASSWORD: secret
    expose:
      - 5432
    healthcheck:
      test: ["CMD-SHELL", "pg_isready"]
      interval: 1s
      timeout: 60s
      retries: 60
      start_period: 80s
  {{.Name}}:
    image: {{.Name}}
    platform: linux/amd64
    depends_on:
      db:
        condition: service_healthy
    links:
      - db
    ports:
      - "8891:8891"
      - "8892:8892"
    environment:
      LOG_LEVEL: debug
      FTL_CONFIG: /root/ftl-project.toml
      FTL_CONTROLLER_DSN: postgres://postgres:secret@db:5432/ftl?sslmode=disable
`

func init() {
	if strings.Contains(boxComposeFile, "\t") {
		panic("tabs in boxComposeFile in cmd_box.go")
	}
}

type boxCmd struct {
	BaseImage   string   `help:"Name of the ftl-box Docker image to use as a base." default:"ftl0/ftl-box:${version}"`
	Parallelism int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Compose     string   `help:"Path to a compose file to generate."`
	Name        string   `arg:"" help:"Name of the project."`
	Dirs        []string `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
}

func (b *boxCmd) Help() string {
	return `
To build a new box with echo and time from examples/go:

	ftl box echo --compose=echo-compose.yml

To run the box:

	docker compose -f echo-compose.yml up --recreate --watch

Interact with the box:

	ftl schema
	ftl ps
	ftl call echo echo '{name:"Alice"}'

Bring the box down:

	docker compose -f echo-compose.yml down --rmi local
	`
}

func (b *boxCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient, projConfig projectconfig.Config) error {
	if len(b.Dirs) == 0 {
		b.Dirs = projConfig.AbsModuleDirs()
	}
	if len(b.Dirs) == 0 {
		return errors.New("no directories specified")
	}
	engine, err := buildengine.New(ctx, client, projConfig.Root(), b.Dirs, buildengine.Parallelism(b.Parallelism))
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
		files = append(files, config.Schema)
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
	err = writeFile(filepath.Join(workDir, "Dockerfile"), boxDockerFile, struct{ BaseImage string }{BaseImage: baseImage})
	if err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}
	err = writeFile(filepath.Join(workDir, "ftl-project.toml"), boxftlProjectFile, nil)
	if err != nil {
		return fmt.Errorf("failed to write ftl-project.toml: %w", err)
	}
	logger.Infof("Building image %s", b.Name)
	err = exec.Command(ctx, log.Debug, workDir, "docker", "build", "-t", b.Name, "--progress=plain", "--platform=linux/amd64", ".").RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	if b.Compose != "" {
		err = writeFile(b.Compose, boxComposeFile, struct {
			Name   string
			GOARCH string
		}{
			Name:   b.Name,
			GOARCH: runtime.GOARCH,
		})
		if err != nil {
			return fmt.Errorf("failed to write compose file: %w", err)
		}
		logger.Infof("Wrote compose file %s", b.Compose)
	}
	return nil
}

func writeFile(path, content string, context any) error {
	t := template.Must(template.New(path).Parse(content))
	w, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer w.Close() //nolint:errcheck
	err = t.Execute(w, context)
	if err != nil {
		return fmt.Errorf("failed to write %q: %w", path, err)
	}
	return nil
}
