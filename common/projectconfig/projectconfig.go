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

type Executables struct {
	FTL string `toml:"ftl"`
}

type Commands struct {
	Startup []string `toml:"startup"`
}

type ConfigAndSecrets struct {
	Config  map[string]*URL `toml:"configuration"`
	Secrets map[string]*URL `toml:"secrets"`
}

type Config struct {
	Global        ConfigAndSecrets            `toml:"global"`
	Modules       map[string]ConfigAndSecrets `toml:"modules"`
	ModuleDirs    []string                    `toml:"module-dirs"`
	ExternalDirs  []string                    `toml:"external-dirs"`
	Executables   Executables                 `toml:"executables"`
	Commands      Commands                    `toml:"commands"`
	FTLMinVersion string                      `toml:"ftl-min-version"`
}

// ConfigPaths returns the computed list of configuration paths to load.
func ConfigPaths(input []string) []string {
	if len(input) > 0 {
		return input
	}
	path := GetDefaultConfigPath()
	_, err := os.Stat(path)
	if err == nil {
		return []string{path}
	}
	return []string{}
}

func GetDefaultConfigPath() string {
	return filepath.Join(internal.GitRoot(""), "ftl-project.toml")
}

// CreateWritableFileIfNonexistent checks the last config file in the list of specified
// paths and creates the file if it does not already exist.
func CreateWritableFileIfNonexistent(ctx context.Context, input []string) error {
	logger := log.FromContext(ctx)

	configPaths := ConfigPaths(input)
	if len(configPaths) == 0 {
		return nil
	}
	path := configPaths[len(configPaths)-1]

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		logger.Warnf("Creating a new project config file at %q because the file does not already exist", path)
		err = Save(path, Config{})
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadConfig(ctx context.Context, input []string) (Config, error) {
	logger := log.FromContext(ctx)
	configPaths := ConfigPaths(input)
	logger.Tracef("Loading config from %s", strings.Join(configPaths, " "))
	config, err := Merge(configPaths...)
	if err != nil {
		return Config{}, err
	}
	if config.FTLMinVersion != "" && !ftl.IsVersionAtLeastMin(ftl.Version, config.FTLMinVersion) {
		return config, fmt.Errorf("FTL version %q predates the minimum version %q", ftl.Version, config.FTLMinVersion)
	}
	return config, nil
}

// LoadWritableConfig loads the last config file in the list of paths, or an empty config if none are found.
func LoadWritableConfig(ctx context.Context, input []string) (Config, error) {
	configPaths := ConfigPaths(input)
	if len(configPaths) == 0 {
		return Config{}, nil
	}
	target := configPaths[len(configPaths)-1]
	log.FromContext(ctx).Tracef("Loading config from %s", target)
	return loadFile(target)
}

// Load project config from a file.
func loadFile(path string) (Config, error) {
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
	return config, nil
}

// Save project config atomically to a file.
func Save(path string, config Config) error {
	w, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return err
	}
	defer os.Remove(w.Name()) //nolint:errcheck
	defer w.Close()           //nolint:errcheck

	enc := toml.NewEncoder(w)
	if err := enc.Encode(config); err != nil {
		return err
	}
	return os.Rename(w.Name(), path)
}
