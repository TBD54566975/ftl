package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	commonruntime "github.com/TBD54566975/ftl/common-runtime"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

type initCmd struct {
	Hermit       bool     `help:"Include Hermit language-specific toolchain binaries in the module." negatable:""`
	Dir          string   `arg:"" help:"Directory to initialize the project in."`
	ExternalDirs []string `help:"Directories of existing external modules."`
	ModuleDirs   []string `help:"Child directories of existing modules."`
	NoGit        bool     `help:"Do not initialize a git repository in the project directory."`
	Startup      string   `help:"Command to run on startup."`
}

func (i initCmd) Run(ctx context.Context) error {
	if i.Dir == "" {
		return fmt.Errorf("directory is required")
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Initializing FTL project in %s", i.Dir)
	if err := scaffold(ctx, i.Hermit, commonruntime.Files(), i.Dir, i); err != nil {
		return err
	}

	if !i.NoGit {
		if err := maybeCreateGitRepo(ctx, i.Dir); err != nil {
			return err
		}
	}

	config := projectconfig.Config{
		ExternalDirs: i.ExternalDirs,
		ModuleDirs:   i.ModuleDirs,
		Commands: projectconfig.Commands{
			Startup: []string{i.Startup},
		},
	}
	return projectconfig.CreateDefault(ctx, config, i.Dir)
}

func maybeCreateGitRepo(ctx context.Context, dir string) error {
	_, hasGitRoot := internal.GitRoot(dir).Get()
	if !hasGitRoot {
		if err := exec.Command(ctx, log.Debug, dir, "git", "init").RunBuffered(ctx); err != nil {
			return err
		}
	}
	if err := updateGitIgnore(dir); err != nil {
		return err
	}
	return nil
}

func updateGitIgnore(dir string) error {
	gitRoot, ok := internal.GitRoot(dir).Get()
	if !ok {
		return nil
	}
	f, err := os.OpenFile(path.Join(gitRoot, ".gitignore"), os.O_RDWR|os.O_CREATE, 0644) //nolint:gosec
	if err != nil {
		return err
	}
	defer f.Close() //nolint:gosec

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "**/_ftl" {
			return nil
		}
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	// append if not already present
	_, err = f.WriteString("**/_ftl\n")
	return err
}
