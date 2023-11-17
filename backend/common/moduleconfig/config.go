package moduleconfig

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/errors"
)

// ModuleConfig is the configuration for an FTL module.
//
// Module config files are currently TOML.
type ModuleConfig struct {
	Language  string   `toml:"language"`
	Module    string   `toml:"module"`
	Build     string   `toml:"build"`
	Deploy    []string `toml:"deploy"`
	DeployDir string   `toml:"deploy-dir"`
	Schema    string   `toml:"schema"`
}

// LoadConfig from a directory.
func LoadConfig(dir string) (ModuleConfig, error) {
	path := filepath.Join(dir, "ftl.toml")
	config := ModuleConfig{}
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return ModuleConfig{}, errors.WithStack(err)
	}
	return config, nil
}
