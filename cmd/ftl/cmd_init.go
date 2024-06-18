package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/TBD54566975/ftl"
	commonruntime "github.com/TBD54566975/ftl/common-runtime"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

type initCmd struct {
	Hermit       bool     `help:"Include Hermit language-specific toolchain binaries." negatable:""`
	Dir          string   `arg:"" help:"Directory to initialize the project in."`
	ExternalDirs []string `help:"Directories of existing external modules."`
	ModuleDirs   []string `help:"Child directories of existing modules."`
	NoGit        bool     `help:"Don't add files to the git repository."`
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

	config := projectconfig.Config{
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

	gitRoot, ok := internal.GitRoot(i.Dir).Get()
	if !i.NoGit && ok {
		logger.Debugf("Updating .gitignore")
		if err := updateGitIgnore(ctx, gitRoot); err != nil {
			return err
		}
		logger.Debugf("Adding files to git")
		if i.Hermit {
			if err := maybeGitAdd(ctx, i.Dir, "bin/*"); err != nil {
				return err
			}
		}
		if err := maybeGitAdd(ctx, i.Dir, "ftl-project.toml"); err != nil {
			return err
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
		if strings.TrimSpace(scanner.Text()) == "**/_ftl" {
			return nil
		}
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	// append if not already present
	if _, err = f.WriteString("**/_ftl\n"); err != nil {
		return err
	}

	// Add .gitignore to git
	return maybeGitAdd(ctx, gitRoot, ".gitignore")
}
