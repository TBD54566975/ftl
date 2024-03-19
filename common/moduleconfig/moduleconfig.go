package moduleconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// ModuleGoConfig is language-specific configuration for Go modules.
type ModuleGoConfig struct{}

// ModuleKotlinConfig is language-specific configuration for Kotlin modules.
type ModuleKotlinConfig struct{}

// ModuleType represents the type of the module.
type ModuleType int

const (
	FTL ModuleType = iota
	ExternalLibrary
)

// ModuleConfig is the configuration for an FTL module.
//
// Module config files are currently TOML.
type ModuleConfig struct {
	// Dir is the root of the module.
	Dir  string     `toml:"-"`
	Type ModuleType `toml:"-"`

	Language string `toml:"language"`
	Realm    string `toml:"realm"`
	Module   string `toml:"module"`
	// Build is the command to build the module.
	Build string `toml:"build"`
	// Deploy is the list of files to deploy relative to the DeployDir.
	Deploy []string `toml:"deploy"`
	// DeployDir is the directory to deploy from, relative to the module directory.
	DeployDir string `toml:"deploy-dir"`
	// Schema is the name of the schema file relative to the DeployDir.
	Schema string `toml:"schema"`
	// Watch is the list of files to watch for changes.
	Watch []string `toml:"watch"`

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
	config.Type = FTL
	config.Dir = dir
	return config, nil
}

func LoadExternalLibraryConfig(dir string) (ModuleConfig, error) {
	config := ModuleConfig{
		Type:   ExternalLibrary,
		Dir:    dir,
		Module: dir,
	}

	goModPath := filepath.Join(dir, "go.mod")
	pomPath := filepath.Join(dir, "pom.xml")
	if _, err := os.Stat(goModPath); err == nil {
		config.Language = "go"
		config.Watch = []string{"**/*.go", "go.mod", "go.sum"}
	} else if !os.IsNotExist(err) {
		return ModuleConfig{}, err
	} else {
		if _, err = os.Stat(pomPath); err == nil {
			config.Language = "kotlin"
			config.Watch = []string{"pom.xml", "src/**", "target/generated-sources"}
		} else if !os.IsNotExist(err) {
			return ModuleConfig{}, err
		}
	}
	if config.Language == "" {
		return ModuleConfig{}, fmt.Errorf("could not autodetect language: no go.mod or pom.xml found in %s", dir)
	}

	return config, nil
}

// AbsDeployDir returns the absolute path to the deploy directory.
func (c ModuleConfig) AbsDeployDir() string {
	return filepath.Join(c.Dir, c.DeployDir)
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
