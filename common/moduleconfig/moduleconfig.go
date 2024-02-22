package moduleconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// ModuleGoConfig is language-specific configuration for Go modules.
type ModuleGoConfig struct{}

// ModuleKotlinConfig is language-specific configuration for Kotlin modules.
type ModuleKotlinConfig struct{}

// ModuleConfig is the configuration for an FTL module.
//
// Module config files are currently TOML.
type ModuleConfig struct {
	Dir string `toml:"-"` // Directory the module config was loaded from.

	Language  string   `toml:"language"`
	Realm     string   `toml:"realm"`
	Module    string   `toml:"module"`
	Build     string   `toml:"build"`
	Deploy    []string `toml:"deploy"`
	DeployDir string   `toml:"deploy-dir"`
	Schema    string   `toml:"schema"`
	Watch     []string `toml:"watch"`

	Go     ModuleGoConfig     `toml:"go,optional"`
	Kotlin ModuleKotlinConfig `toml:"kotlin,optional"`
}

// LoadModuleConfig from a directory.
func LoadModuleConfig(dir string) (ModuleConfig, error) {
	path := filepath.Join(dir, "ftl.toml")
	config := ModuleConfig{}
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return ModuleConfig{}, err
	}
	if err := setConfigDefaults(dir, &config); err != nil {
		return config, fmt.Errorf("%s: %w", path, err)
	}
	config.Dir = dir
	return config, nil
}

func setConfigDefaults(moduleDir string, config *ModuleConfig) error {
	if config.Realm == "" {
		config.Realm = "home"
	}
	if config.Schema == "" {
		config.Schema = "schema.pb"
	}
	switch config.Language {
	case "kotlin":
		if config.Build == "" {
			config.Build = "mvn -B compile"
		}
		if config.DeployDir == "" {
			config.DeployDir = "target"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main", "classes", "dependency", "classpath.txt"}
		}
		if len(config.Watch) == 0 {
			config.Watch = []string{"pom.xml", "src/**", "target/generated-sources"}
		}

	case "go":
		if config.DeployDir == "" {
			config.DeployDir = "_ftl"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main"}
		}
		if len(config.Watch) == 0 {
			config.Watch = []string{"**/*.go", "go.mod", "go.sum"}
		}
	}

	// Do some validation.
	if !isBeneath(moduleDir, config.DeployDir) {
		return fmt.Errorf("deploy-dir must be relative to the module directory")
	}
	for _, deploy := range config.Deploy {
		if !isBeneath(moduleDir, deploy) {
			return fmt.Errorf("deploy files must be relative to the module directory")
		}
	}
	for _, watch := range config.Watch {
		if !isBeneath(moduleDir, watch) {
			return fmt.Errorf("watch files must be relative to the module directory")
		}
	}
	return nil
}

func isBeneath(moduleDir, path string) bool {
	resolved := filepath.Clean(filepath.Join(moduleDir, path))
	return strings.HasPrefix(resolved, strings.TrimSuffix(moduleDir, "/")+"/")
}
