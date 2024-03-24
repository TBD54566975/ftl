package projectconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/log"
)

type ConfigAndSecrets struct {
	Config  map[string]*URL `toml:"configuration"`
	Secrets map[string]*URL `toml:"secrets"`
}

type Config struct {
	Global       ConfigAndSecrets            `toml:"global"`
	Modules      map[string]ConfigAndSecrets `toml:"modules"`
	ModuleDirs   []string                    `toml:"module-dirs"`
	ExternalDirs []string                    `toml:"external-dirs"`
}

// ConfigPaths returns the computed list of configuration paths to load.
func ConfigPaths(input []string) []string {
	if len(input) > 0 {
		return input
	}
	path := filepath.Join(internal.GitRoot(""), "ftl-project.toml")
	_, err := os.Stat(path)
	if err == nil {
		return []string{path}
	}
	return []string{}
}

func LoadConfig(ctx context.Context, input []string) (Config, error) {
	logger := log.FromContext(ctx)
	configPaths := ConfigPaths(input)
	logger.Tracef("Loading config from %s", strings.Join(configPaths, " "))
	config, err := Merge(configPaths...)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

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
