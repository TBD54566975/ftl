package main

import (
	"archive/zip"
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/block/scaffolder"

	"github.com/block/ftl"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/providers"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/profiles"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/projectinit"
)

//go:embed dependency-versions.txt
var userHermitPackages string

type initCmd struct {
	Name        string   `arg:"" help:"Name of the project."`
	Hermit      bool     `help:"Include Hermit language-specific toolchain binaries." negatable:"" default:"true"`
	Dir         string   `arg:"" help:"Directory to initialize the project in." default:"." required:""`
	ModuleDirs  []string `help:"Child directories of existing modules."`
	ModuleRoots []string `help:"Root directories of existing modules."`
	NoGit       bool     `help:"Don't add files to the git repository."`
	Startup     string   `help:"Command to run on startup."`
}

func (i initCmd) Run(
	ctx context.Context,
	logger *log.Logger,
	configRegistry *providers.Registry[configuration.Configuration],
	secretsRegistry *providers.Registry[configuration.Secrets],
) error {
	logger.Debugf("Initializing FTL project in %s", i.Dir)
	if err := scaffold(ctx, i.Hermit, projectinit.Files(), i.Dir, i); err != nil {
		return err
	}

	config := projectconfig.Config{
		Name:          i.Name,
		Hermit:        i.Hermit,
		NoGit:         i.NoGit,
		FTLMinVersion: ftl.Version,
		ModuleDirs:    i.ModuleDirs,
		Commands: projectconfig.Commands{
			Startup: []string{i.Startup},
		},
	}
	if err := projectconfig.Create(ctx, config, i.Dir); err != nil {
		return err
	}

	_, err := profiles.Init(profiles.ProjectConfig{
		Realm:         i.Name,
		FTLMinVersion: ftl.Version,
		ModuleRoots:   i.ModuleRoots,
		NoGit:         i.NoGit,
		Root:          i.Dir,
	}, secretsRegistry, configRegistry)
	if err != nil {
		return fmt.Errorf("initialize project: %w", err)
	}
	if i.Hermit {
		if err := installHermitFTL(ctx, i.Dir); err != nil {
			return fmt.Errorf("initialize Hermit FTL: %w", err)
		}
	}

	if !i.NoGit {
		err := maybeGitInit(ctx, i.Dir)
		if err != nil {
			return fmt.Errorf("running git init: %w", err)
		}
		logger.Debugf("Updating .gitignore")
		if err := updateGitIgnore(ctx, i.Dir); err != nil {
			return fmt.Errorf("update .gitignore: %w", err)
		}
		if err := maybeGitAdd(ctx, i.Dir, ".ftl-project"); err != nil {
			return fmt.Errorf("git add .ftl-project: %w", err)
		}
		if err := maybeGitAdd(ctx, i.Dir, "ftl-project.toml"); err != nil {
			return fmt.Errorf("git add ftl-project.toml: %w", err)
		}
		if err := maybeGitAdd(ctx, i.Dir, "README.md"); err != nil {
			return fmt.Errorf("git add README.md: %w", err)
		}
		if i.Hermit {
			if err := maybeGitAdd(ctx, i.Dir, "bin"); err != nil {
				return fmt.Errorf("git add bin: %w", err)
			}
		}
	}
	return nil
}

func maybeGitAdd(ctx context.Context, dir string, paths ...string) error {
	args := append([]string{"add"}, paths...)
	if err := exec.Command(ctx, log.Debug, dir, "git", args...).RunBuffered(ctx); err != nil {
		return err
	}
	return nil
}

func installHermitFTL(ctx context.Context, dir string) error {
	for _, install := range strings.Split(userHermitPackages, "\n") {
		if install == "" {
			continue
		}
		args := []string{"install", install}
		if err := exec.Command(ctx, log.Debug, dir, "./bin/hermit", args...).RunBuffered(ctx); err != nil {
			return fmt.Errorf("unable to install hermit package %s %w", install, err)
		}
	}
	args := []string{"install", "ftl@" + ftl.Version}
	if err := exec.Command(ctx, log.Debug, dir, "./bin/hermit", args...).RunBuffered(ctx); err != nil {
		return err
	}
	return nil
}

func maybeGitInit(ctx context.Context, dir string) error {
	args := []string{"init"}
	if err := exec.Command(ctx, log.Debug, dir, "git", args...).RunBuffered(ctx); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	return nil
}

func updateGitIgnore(ctx context.Context, gitRoot string) error {
	f, err := os.OpenFile(path.Join(gitRoot, ".gitignore"), os.O_RDWR|os.O_CREATE, 0644) //nolint:gosec
	if err != nil {
		return err
	}
	defer f.Close() //nolint:gosec

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "**/.ftl" {
			return nil
		}
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	// append if not already present
	if _, err = f.WriteString("**/.ftl\n"); err != nil {
		return err
	}

	// Add .gitignore to git
	return maybeGitAdd(ctx, gitRoot, ".gitignore")
}

func scaffold(ctx context.Context, includeBinDir bool, source *zip.Reader, destination string, sctx any, options ...scaffolder.Option) error {
	logger := log.FromContext(ctx)
	opts := []scaffolder.Option{scaffolder.Exclude("^go.mod$")}
	if !includeBinDir {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}
	opts = append(opts, options...)
	if err := internal.ScaffoldZip(source, destination, sctx, opts...); err != nil {
		return fmt.Errorf("failed to scaffold: %w", err)
	}
	return nil
}
