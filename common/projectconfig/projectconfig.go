package projectconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
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

// Load project config from a file.
func Load(path string) (Config, error) {
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
