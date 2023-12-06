package moduleconfig

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
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
	Watch     []string `toml:"watch"`
}

// LoadConfig from a directory.
func LoadConfig(dir string) (ModuleConfig, error) {
	path := filepath.Join(dir, "ftl.toml")
	config := ModuleConfig{}
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return ModuleConfig{}, err
	}
	setConfigDefaults(&config)
	return config, nil
}

func setConfigDefaults(config *ModuleConfig) {
	switch config.Language {
	case "kotlin":
		if config.Build == "" {
			config.Build = "mvn --batch-mode compile"
		}
		if config.DeployDir == "" {
			config.DeployDir = "target"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main", "classes", "dependency", "classpath.txt"}
		}
		if config.Schema == "" {
			config.Schema = "schema.pb"
		}
		if len(config.Watch) == 0 {
			config.Watch = []string{"pom.xml", "src/**", "target/generated-sources"}
		}
	case "go":
		if config.DeployDir == "" {
			config.DeployDir = "build"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main", "schema.pb"}
		}
	}
}
