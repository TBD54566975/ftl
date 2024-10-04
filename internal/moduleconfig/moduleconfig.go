package moduleconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-viper/mapstructure/v2"

	"github.com/TBD54566975/ftl/internal/slices"
)

// ModuleConfig is the configuration for an FTL module.
//
// Module config files are currently TOML.
type ModuleConfig struct {
	// Dir is the absolute path to the root of the module.
	Dir string `toml:"-"`

	Language string `toml:"language"`
	Realm    string `toml:"realm"`
	Module   string `toml:"module"`
	// Build is the command to build the module.
	Build string `toml:"build"`
	// Deploy is the list of files to deploy relative to the DeployDir.
	Deploy []string `toml:"deploy"`
	// DeployDir is the directory to deploy from, relative to the module directory.
	DeployDir string `toml:"deploy-dir"`
	// GeneratedSchemaDir is the directory to generate protobuf schema files into. These can be picked up by language specific build tools
	GeneratedSchemaDir string `toml:"generated-schema-dir"`
	// Errors is the name of the error file relative to the DeployDir.
	Errors string `toml:"errors"`
	// Watch is the list of files to watch for changes.
	Watch []string `toml:"watch"`

	// LanguageConfig is a map of language specific configuration.
	// It is saved in the toml with the value of Language as the key.
	LanguageConfig map[string]any `toml:"-"`
}

func (c *ModuleConfig) UnmarshalTOML(data []byte) error {
	return nil
}

// AbsModuleConfig is a ModuleConfig with all paths made absolute.
//
// This is a type alias to prevent accidental use of the wrong type.
type AbsModuleConfig ModuleConfig

// UnvalidatedModuleConfig is a ModuleConfig that has been loaded from disk
// but has not had it's defaults set or been validated.
type UnvalidatedModuleConfig ModuleConfig

type CustomDefaults struct {
	Build              string
	Deploy             []string
	DeployDir          string
	GeneratedSchemaDir string
	Errors             string
	Watch              []string
}

// LoadModuleConfig from a directory.
func LoadModuleConfig(dir string) (UnvalidatedModuleConfig, error) {
	path := filepath.Join(dir, "ftl.toml")

	// Parse toml into generic map so that we can capture language config with a dynamic key
	raw := map[string]any{}
	_, err := toml.DecodeFile(path, &raw)
	if err != nil {
		return UnvalidatedModuleConfig{}, fmt.Errorf("could not parse module toml: %w", err)
	}

	// Decode the generic map into a module config
	config := UnvalidatedModuleConfig{
		Dir: dir,
	}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: false,
		TagName:     "toml",
		Result:      &config,
	})
	if err != nil {
		return UnvalidatedModuleConfig{}, fmt.Errorf("could not parse contents of module toml: %w", err)
	}
	err = mapDecoder.Decode(raw)
	if err != nil {
		return UnvalidatedModuleConfig{}, fmt.Errorf("could not parse contents of module toml: %w", err)
	}
	// Decode language config
	rawLangConfig, ok := raw[config.Language]
	if ok {
		langConfig, ok := rawLangConfig.(map[string]any)
		if !ok {
			return UnvalidatedModuleConfig{}, fmt.Errorf("language config for %q is not a map", config.Language)
		}
		config.LanguageConfig = langConfig
	}
	return config, nil
}

func (c ModuleConfig) String() string {
	return fmt.Sprintf("%s (%s)", c.Module, c.Dir)
}

// Abs creates a clone of ModuleConfig with all paths made absolute.
//
// This function will panic if any paths are not beneath the module directory.
// This should never happen under normal use, as LoadModuleConfig performs this
// validation separately. This is just a sanity check.
func (c ModuleConfig) Abs() AbsModuleConfig {
	clone := c
	clone.Dir = filepath.Clean(clone.Dir)
	clone.DeployDir = filepath.Clean(filepath.Join(clone.Dir, clone.DeployDir))
	if !strings.HasPrefix(clone.DeployDir, clone.Dir) {
		panic(fmt.Sprintf("deploy-dir %q is not beneath module directory %q", clone.DeployDir, clone.Dir))
	}
	if clone.GeneratedSchemaDir != "" {
		clone.GeneratedSchemaDir = filepath.Clean(filepath.Join(clone.Dir, clone.GeneratedSchemaDir))
		if !strings.HasPrefix(clone.GeneratedSchemaDir, clone.Dir) {
			panic(fmt.Sprintf("generated-schema-dir %q is not beneath module directory %q", clone.GeneratedSchemaDir, clone.Dir))
		}
	}
	clone.Errors = filepath.Clean(filepath.Join(clone.DeployDir, clone.Errors))
	if !strings.HasPrefix(clone.Errors, clone.DeployDir) {
		panic(fmt.Sprintf("errors %q is not beneath deploy directory %q", clone.Errors, clone.DeployDir))
	}
	clone.Deploy = slices.Map(clone.Deploy, func(p string) string {
		out := filepath.Clean(filepath.Join(clone.DeployDir, p))
		if !strings.HasPrefix(out, clone.DeployDir) {
			panic(fmt.Sprintf("deploy path %q is not beneath deploy directory %q", out, clone.DeployDir))
		}
		return out
	})
	// Watch paths are allowed to be outside the deploy directory.
	clone.Watch = slices.Map(clone.Watch, func(p string) string {
		return filepath.Clean(filepath.Join(clone.Dir, p))
	})
	return AbsModuleConfig(clone)
}

// configureDefaults defaults sets values for empty fields and validates the config.
// It involves standard defaults for Real and Errors fields, and also looks at CustomDefaults for
// defaulting other fields.
func (c UnvalidatedModuleConfig) DefaultAndValidate(customDefaults CustomDefaults) (ModuleConfig, error) {
	if c.Realm == "" {
		c.Realm = "home"
	}
	if c.Errors == "" {
		c.Errors = "errors.pb"
	}

	// Custom defaults
	if c.Build == "" {
		c.Build = customDefaults.Build
	}
	if c.DeployDir == "" {
		c.DeployDir = customDefaults.DeployDir
	}
	if c.Deploy == nil {
		c.Deploy = customDefaults.Deploy
	}
	if c.GeneratedSchemaDir == "" {
		c.GeneratedSchemaDir = customDefaults.GeneratedSchemaDir
	}
	if c.Watch == nil {
		c.Watch = customDefaults.Watch
	}

	// Validate
	if !isBeneath(c.Dir, c.DeployDir) {
		return ModuleConfig{}, fmt.Errorf("deploy-dir %s must be relative to the module directory %s", c.DeployDir, c.Dir)
	}
	for _, deploy := range c.Deploy {
		if !isBeneath(c.Dir, deploy) {
			return ModuleConfig{}, fmt.Errorf("deploy %s files must be relative to the module directory %s", deploy, c.Dir)
		}
	}
	c.Deploy = slices.Sort(c.Deploy)
	c.Watch = slices.Sort(c.Watch)
	return ModuleConfig(c), nil
}

func isBeneath(moduleDir, path string) bool {
	resolved := filepath.Clean(filepath.Join(moduleDir, path))
	return strings.HasPrefix(resolved, strings.TrimSuffix(moduleDir, "/")+"/")
}

func (c ModuleConfig) Schema() string {
	return "schema.pb"
}

func (c AbsModuleConfig) Schema() string {
	return filepath.Join(c.DeployDir, "schema.pb")
}
