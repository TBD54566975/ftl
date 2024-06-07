package projectconfig

import (
	"path/filepath"
)

// Merge configuration files.
//
// Config is merged left to right, with later files taking precedence over earlier files.
func Merge(paths ...string) (Config, error) {
	config := Config{}
	for _, path := range paths {
		partial, err := loadFile(path)
		if err != nil {
			return config, err
		}
		config = merge(config, partial)

		// Make module-dirs absolute to mimic the behavior of the CLI
		config.absModuleDirs = []string{}
		for _, dir := range config.ModuleDirs {
			if !filepath.IsAbs(dir) {
				absDir := filepath.Join(filepath.Dir(path), dir)
				config.absModuleDirs = append(config.absModuleDirs, absDir)
			} else {
				config.absModuleDirs = append(config.absModuleDirs, dir)
			}
		}

		// If no module-dirs are defined, default to the directory of the config file
		if len(config.absModuleDirs) == 0 {
			config.absModuleDirs = []string{filepath.Dir(path)}
		}
	}

	return config, nil
}

func merge(a, b Config) Config {
	a = mergeRootKeys(a, b)
	a.Global = mergeConfigAndSecrets(a.Global, b.Global)
	for k, v := range b.Modules {
		if a.Modules == nil {
			a.Modules = map[string]ConfigAndSecrets{}
		}
		a.Modules[k] = mergeConfigAndSecrets(a.Modules[k], v)
	}
	return a
}

func mergeConfigAndSecrets(a, b ConfigAndSecrets) ConfigAndSecrets {
	for k, v := range b.Config {
		if a.Config == nil {
			a.Config = map[string]*URL{}
		}
		a.Config[k] = v
	}
	for k, v := range b.Secrets {
		if a.Secrets == nil {
			a.Secrets = map[string]*URL{}
		}
		a.Secrets[k] = v
	}
	return a
}

func mergeRootKeys(a, b Config) Config {
	if b.ModuleDirs != nil {
		a.ModuleDirs = b.ModuleDirs
	}
	if b.ExternalDirs != nil {
		a.ExternalDirs = b.ExternalDirs
	}
	if len(b.Commands.Startup) > 0 {
		a.Commands.Startup = b.Commands.Startup
	}
	if b.FTLMinVersion != "" {
		a.FTLMinVersion = b.FTLMinVersion
	}
	return a
}
