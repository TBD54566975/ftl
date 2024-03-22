package projectconfig

// Merge configuration files.
//
// Config is merged left to right, with later files taking precedence over earlier files.
func Merge(paths ...string) (Config, error) {
	config := Config{}
	for _, path := range paths {
		partial, err := Load(path)
		if err != nil {
			return config, err
		}
		config = merge(config, partial)
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
	if a.ModuleDirs == nil {
		a.ModuleDirs = b.ModuleDirs
	}
	if a.ExternalDirs == nil {
		a.ExternalDirs = b.ExternalDirs
	}
	return a
}
