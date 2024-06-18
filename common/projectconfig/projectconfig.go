package projectconfig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/types/optional"

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

// DefaultConfigPath returns the absolute default path for the project config file, if possible.
//
// The default path is determined by the FTL_CONFIG environment variable, if set, or by the presence of a Git
// repository. If the Git repository is found, the default path is the root of the repository with the filename
// "ftl-project.toml".
func DefaultConfigPath() optional.Option[string] {
	if envar, ok := os.LookupEnv("FTL_CONFIG"); ok {
		absPath, err := filepath.Abs(envar)
		if err != nil {
			return optional.None[string]()
		}
		return optional.Some(absPath)
	}
	dir, err := os.Getwd()
	if err != nil {
		return optional.None[string]()
	}
	// Find the first ftl-project.toml file in the parent directories, up until the gitroot.
	root, ok := internal.GitRoot(dir).Get()
	if !ok {
		root = "/"
	}
	for dir != root && dir != "." {
		path := filepath.Join(dir, "ftl-project.toml")
		_, err := os.Stat(path)
		if err == nil {
			return optional.Some(path)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return optional.None[string]()
		}
		dir = filepath.Dir(dir)
	}
	return optional.Some(filepath.Join(dir, "ftl-project.toml"))
}

// MaybeCreateDefault creates the ftl-project.toml file in the Git root if it
// does not already exist.
func MaybeCreateDefault(ctx context.Context) error {
	logger := log.FromContext(ctx)
	path, ok := DefaultConfigPath().Get()
	if !ok {
		logger.Warnf("Failed to find Git root, so cannot verify whether an ftl-project.toml file exists there")
		return nil
	}
	_, err := os.Stat(path)
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
		maybePath, ok := DefaultConfigPath().Get()
		if !ok {
			return Config{}, nil
		}
		path = maybePath
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return Config{}, err
	}
	config := Config{}
	md, err := toml.DecodeFile(path, &config)
	if err != nil {
		return Config{}, err
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
