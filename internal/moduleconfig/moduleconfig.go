package moduleconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/types/optional"
	"github.com/go-viper/mapstructure/v2"

	"github.com/block/ftl/common/slices"
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
	// Build is the command to build the module in dev mode.
	DevModeBuild string `toml:"dev-mode-build"`
	// BuildLock is file lock path to prevent concurrent builds of a module.
	BuildLock string `toml:"build-lock"`
	// DeployDir is the directory to deploy from, relative to the module directory.
	DeployDir string `toml:"deploy-dir"`
	// Watch is the list of files to watch for changes.
	Watch []string `toml:"watch"`

	// LanguageConfig is a map of language specific configuration.
	// It is saved in the toml with the value of Language as the key.
	LanguageConfig map[string]any `toml:"-"`
	// SQLMigrationDirectory is the directory to look for SQL migrations.
	SQLMigrationDirectory string `toml:"sql-migration-directory"`
}

func (c *ModuleConfig) UnmarshalTOML(data []byte) error {
	return nil
}

// AbsModuleConfig is a ModuleConfig with all paths made absolute.
//
// This is a type alias to prevent accidental use of the wrong type.
type AbsModuleConfig ModuleConfig

// UnvalidatedModuleConfig is a ModuleConfig that holds only the values read from the toml file.
//
// It has not had it's defaults set or been validated, so values may be empty or invalid.
// Use FillDefaultsAndValidate() to get a ModuleConfig.
type UnvalidatedModuleConfig ModuleConfig

type CustomDefaults struct {
	DeployDir          string
	Watch              []string
	BuildLock          optional.Option[string]
	Build              optional.Option[string]
	DevModeBuild       optional.Option[string]
	GeneratedSchemaDir optional.Option[string]

	// only the root keys in LanguageConfig are used to find missing values that can be defaulted
	LanguageConfig map[string]any `toml:"-"`

	// SQLMigrationDir is the directory to look for SQL migrations.
	SQLMigrationDir string
}

// LoadConfig from a directory.
// This returns only the values found in the toml file. To get the full config with defaults and validation, use FillDefaultsAndValidate.
func LoadConfig(dir string) (UnvalidatedModuleConfig, error) {
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
	dir, err := filepath.Abs(filepath.Clean(clone.Dir))
	if err != nil {
		panic(fmt.Sprintf("module dir %q can not be made absolute", c.Dir))
	}
	clone.Dir = dir
	clone.DeployDir = filepath.Clean(filepath.Join(clone.Dir, clone.DeployDir))
	clone.SQLMigrationDirectory = filepath.Clean(filepath.Join(clone.Dir, clone.SQLMigrationDirectory))
	if !strings.HasPrefix(clone.DeployDir, clone.Dir) {
		panic(fmt.Sprintf("deploy-dir %q is not beneath module directory %q", clone.DeployDir, clone.Dir))
	}
	clone.BuildLock = filepath.Clean(filepath.Join(clone.Dir, clone.BuildLock))
	// Watch paths are allowed to be outside the deploy directory.
	clone.Watch = slices.Map(clone.Watch, func(p string) string {
		return filepath.Clean(filepath.Join(clone.Dir, p))
	})
	return AbsModuleConfig(clone)
}

// FillDefaultsAndValidate sets values for empty fields and validates the config.
// It involves standard defaults for Real and Errors fields, and also looks at CustomDefaults for
// defaulting other fields.
func (c UnvalidatedModuleConfig) FillDefaultsAndValidate(customDefaults CustomDefaults) (ModuleConfig, error) {
	if c.Realm == "" {
		c.Realm = "home"
	}

	// Custom defaults
	if defaultValue, ok := customDefaults.Build.Get(); ok && c.Build == "" {
		c.Build = defaultValue
	}
	if defaultValue, ok := customDefaults.DevModeBuild.Get(); ok && c.DevModeBuild == "" {
		c.DevModeBuild = defaultValue
	}
	if c.BuildLock == "" {
		if defaultValue, ok := customDefaults.BuildLock.Get(); ok {
			c.BuildLock = defaultValue
		} else {
			c.BuildLock = ".ftl.lock"
		}
	}
	if c.DeployDir == "" {
		c.DeployDir = customDefaults.DeployDir
	}
	if c.SQLMigrationDirectory == "" {
		c.SQLMigrationDirectory = customDefaults.SQLMigrationDir

	}
	if c.Watch == nil {
		c.Watch = customDefaults.Watch
	}

	// Find any missing keys in LanguageConfig that can be defaulted
	if c.LanguageConfig == nil && customDefaults.LanguageConfig != nil {
		c.LanguageConfig = map[string]any{}
	}
	for k, v := range customDefaults.LanguageConfig {
		if _, ok := c.LanguageConfig[k]; !ok {
			c.LanguageConfig[k] = v
		}
	}

	// Validate
	if c.DeployDir == "" {
		return ModuleConfig{}, fmt.Errorf("no deploy directory configured")
	}
	if c.BuildLock == "" {
		return ModuleConfig{}, fmt.Errorf("no build lock path configured")
	}
	if !isBeneath(c.Dir, c.DeployDir) {
		return ModuleConfig{}, fmt.Errorf("deploy-dir %s must be relative to the module directory %s", c.DeployDir, c.Dir)
	}
	c.Watch = slices.Sort(c.Watch)
	return ModuleConfig(c), nil
}

func isBeneath(moduleDir, path string) bool {
	resolved := filepath.Clean(filepath.Join(moduleDir, path))
	return strings.HasPrefix(resolved, strings.TrimSuffix(moduleDir, "/")+"/")
}
