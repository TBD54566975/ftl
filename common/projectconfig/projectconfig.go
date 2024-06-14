package projectconfig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/log"
)

type Commands struct {
	Startup []string `toml:"startup"`
}

type ConfigAndSecrets struct {
	Config  map[string]*URL `toml:"configuration"`
	Secrets map[string]*URL `toml:"secrets"`
}

type Config struct {
	// Path to the config file.
	Path string `toml:"-"`

	Global        ConfigAndSecrets            `toml:"global"`
	Modules       map[string]ConfigAndSecrets `toml:"modules"`
	ModuleDirs    []string                    `toml:"module-dirs"`
	ExternalDirs  []string                    `toml:"external-dirs"`
	Commands      Commands                    `toml:"commands"`
	FTLMinVersion string                      `toml:"ftl-min-version"`
}

// Root directory of the project.
func (c Config) Root() string {
	if !filepath.IsAbs(c.Path) {
		panic(fmt.Errorf("project config path must be absolute: %s", c.Path))
	}
	return filepath.Dir(c.Path)
}

// AbsModuleDirs returns the absolute path for the module-dirs field from the ftl-project.toml, unless
// that is not defined, in which case it defaults to the root directory.
func (c Config) AbsModuleDirs() []string {
	if len(c.ModuleDirs) == 0 {
		return []string{filepath.Dir(c.Path)}
	}
	root := c.Root()
	absDirs := make([]string, len(c.ModuleDirs))
	for i, dir := range c.ModuleDirs {
		cleaned := filepath.Clean(filepath.Join(root, dir))
		if !strings.HasPrefix(cleaned, root) {
			panic(fmt.Errorf("module-dirs path %q is not within the project root %q", dir, root))
		}
		absDirs[i] = cleaned
	}
	return absDirs
}

// The directory the binary was executed from.
var startDir = func() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}()

// DefaultConfigPath returns the absolute default path for the ftl-project.toml
// file.
//
// The semantics are the following:
//
//  1. If the `FTL_CONFIG` environment variable is set, it is used directly.
//  2. Otherwise the parent directories are searched for an existing file.
//  3. If not found and a git root is present, ${GIT_ROOT}/ftl-project.toml is returned.
//  4. Finally, the current directory is used.
func DefaultConfigPath() (string, error) {
	// First try the FTL_CONFIG environment variable.
	if envar, ok := os.LookupEnv("FTL_CONFIG"); ok {
		absPath, err := filepath.Abs(envar)
		if err != nil {
			return "", fmt.Errorf("failed to resolve FTL_CONFIG path %q: %w", envar, err)
		}
		return absPath, nil
	}

	// Next, try to find the first ftl-project.toml file in parent directories.
	//
	// This is used to support the case where the config is in a subdirectory of the project.
	for dir := startDir; dir != "/" && dir != "."; dir = filepath.Dir(dir) {
		path := filepath.Join(dir, "ftl-project.toml")
		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to check for ftl-project.toml in %q: %w", dir, err)
		}
		dir = filepath.Dir(dir)
	}

	// Default to the git root if found.
	if git, ok := internal.GitRoot(startDir).Get(); ok {
		return filepath.Join(git, "ftl-project.toml"), nil
	}

	// Finally, default to the current directory.
	return filepath.Join(startDir, "ftl-project.toml"), nil
}

// MaybeCreateDefault creates the ftl-project.toml file in the Git root if it
// does not already exist.
func MaybeCreateDefault(ctx context.Context) error {
	logger := log.FromContext(ctx)
	path, err := DefaultConfigPath()
	if err != nil {
		return err
	}
	_, err = os.Stat(path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	logger.Debugf("Creating a new project config file at %q", path)
	return Save(Config{Path: path})
}

// LoadOrCreate loads or creates the given configuration file.
func LoadOrCreate(ctx context.Context, target string) (Config, error) {
	logger := log.FromContext(ctx)
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		logger.Debugf("Creating a new project config file at %q", target)
		config := Config{Path: target}
		return config, Save(config)
	}

	log.FromContext(ctx).Tracef("Loading config from %s", target)
	return Load(ctx, target)
}

// Load project config from a file.
func Load(ctx context.Context, path string) (Config, error) {
	if path == "" {
		defaultPath, err := DefaultConfigPath()
		if err != nil {
			return Config{}, err
		}
		path = defaultPath
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to resolve config path %q: %w", path, err)
	}
	config := Config{}
	md, err := toml.DecodeFile(path, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to decode %q: %w", path, err)
	}
	if len(md.Undecoded()) > 0 {
		keys := make([]string, len(md.Undecoded()))
		for i, key := range md.Undecoded() {
			keys[i] = key.String()
		}
		return Config{}, fmt.Errorf("unknown configuration keys: %s", strings.Join(keys, ", "))
	}

	if config.FTLMinVersion != "" && !ftl.IsVersionAtLeastMin(ftl.Version, config.FTLMinVersion) {
		return config, fmt.Errorf("FTL version %q predates the minimum version %q", ftl.Version, config.FTLMinVersion)
	}
	config.Path = path

	for _, dir := range config.ModuleDirs {
		absDir := filepath.Clean(filepath.Join(config.Root(), dir))
		if !strings.HasPrefix(absDir, config.Root()) {
			return Config{}, fmt.Errorf("module-dirs path %q is not within the project root %q", dir, config.Root())
		}
	}

	return config, nil
}

// Save project config to its file atomically.
func Save(config Config) error {
	if config.Path == "" {
		return fmt.Errorf("project config path must be set")
	}
	if !filepath.IsAbs(config.Path) {
		panic(fmt.Errorf("project config path must be absolute: %s", config.Path))
	}
	w, err := os.CreateTemp(filepath.Dir(config.Path), filepath.Base(config.Path))
	if err != nil {
		return err
	}
	defer os.Remove(w.Name()) //nolint:errcheck
	defer w.Close()           //nolint:errcheck

	enc := toml.NewEncoder(w)
	if err := enc.Encode(config); err != nil {
		return err
	}
	return os.Rename(w.Name(), config.Path)
}
