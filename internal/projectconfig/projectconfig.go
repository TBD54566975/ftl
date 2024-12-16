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

	"github.com/block/ftl"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/log"
)

type Commands struct {
	Startup []string `toml:"startup"`
}

type ConfigAndSecrets struct {
	Config  map[string]*URL `toml:"configuration"`
	Secrets map[string]*URL `toml:"secrets"`
}

type Config struct {
	// Path to the config file populated on load.
	Path string `toml:"-"`

	Name                  string                      `toml:"name,omitempty"`
	Global                ConfigAndSecrets            `toml:"global,omitempty"`
	SecretsProvider       configuration.ProviderKey   `toml:"secrets-provider,omitempty"`
	ConfigProvider        configuration.ProviderKey   `toml:"config-provider,omitempty"`
	Modules               map[string]ConfigAndSecrets `toml:"modules,omitempty"`
	ModuleDirs            []string                    `toml:"module-dirs,omitempty"`
	Commands              Commands                    `toml:"commands,omitempty"`
	FTLMinVersion         string                      `toml:"ftl-min-version,omitempty"`
	Hermit                bool                        `toml:"hermit,omitempty"`
	NoGit                 bool                        `toml:"no-git,omitempty"`
	DisableIDEIntegration bool                        `toml:"disable-ide-integration,omitempty"`
}

// Root directory of the project.
func (c Config) Root() string {
	if !filepath.IsAbs(c.Path) {
		panic(fmt.Errorf("project config path must be absolute: %s", c.Path))
	}
	return filepath.Dir(c.Path)
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.SecretsProvider == "" {
		c.SecretsProvider = "inline"
	}
	if c.ConfigProvider == "" {
		c.ConfigProvider = "inline"
	}
	if c.Name == "" {
		return fmt.Errorf("project name is required: %s", c.Path)
	}
	if strings.Contains(c.Name, " ") {
		return fmt.Errorf("project name %q includes spaces: %s", c.Name, c.Path)
	}
	if c.FTLMinVersion != "" && !ftl.IsVersionAtLeastMin(ftl.Version, c.FTLMinVersion) {
		return fmt.Errorf("FTL version %q predates the minimum version %q", ftl.Version, c.FTLMinVersion)
	}
	for _, dir := range c.ModuleDirs {
		absDir := filepath.Clean(filepath.Join(c.Root(), dir))
		if !strings.HasPrefix(absDir, c.Root()) {
			return fmt.Errorf("module-dirs path %q is not within the project root %q", dir, c.Root())
		}
	}
	return nil
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

// Create creates the ftl-project.toml file with the given Config into dir.
func Create(ctx context.Context, config Config, dir string) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("project config: %w", err)
	}
	logger := log.FromContext(ctx)
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	path = filepath.Join(path, "ftl-project.toml")
	_, err = os.Stat(path)
	if err == nil {
		return fmt.Errorf("project config file already exists at %q", path)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	logger.Debugf("Creating a new project config file at %q", path)
	config.Path = path
	return Save(config)
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
	config.Path = path

	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("%s: %w", path, err)
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

// SchemaPath returns the path to the schema file for the given module.
func (c Config) SchemaPath(module string) string {
	return filepath.Join(c.Root(), ".ftl", "schemas", module+".pb")
}
