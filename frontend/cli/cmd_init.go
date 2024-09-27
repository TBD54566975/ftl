package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/profiles"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/projectinit"
)

type initCmd struct {
	Name        string   `arg:"" help:"Name of the project."`
	Hermit      bool     `help:"Include Hermit language-specific toolchain binaries." negatable:""`
	Dir         string   `arg:"" help:"Directory to initialize the project in." default:"${gitroot}" required:""`
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

	if !i.NoGit {
		logger.Debugf("Updating .gitignore")
		if err := updateGitIgnore(ctx, i.Dir); err != nil {
			return fmt.Errorf("update .gitignore: %w", err)
		}
		if err := maybeGitAdd(ctx, i.Dir, ".ftl-project"); err != nil {
			return fmt.Errorf("git add .ftl-project: %w", err)
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
