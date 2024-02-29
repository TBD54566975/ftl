package configuration

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	pc "github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal"
)

type FromConfigOrSecrets interface {
	get(config pc.ConfigAndSecrets) map[string]*pc.URL
	set(config *pc.ConfigAndSecrets, mapping map[string]*pc.URL)
}

type FromConfig struct{}

func (f FromConfig) get(config pc.ConfigAndSecrets) map[string]*pc.URL { return config.Config }

func (f FromConfig) set(config *pc.ConfigAndSecrets, mapping map[string]*pc.URL) {
	config.Config = mapping
}

type FromSecrets struct{}

func (f FromSecrets) get(config pc.ConfigAndSecrets) map[string]*pc.URL { return config.Secrets }

func (f FromSecrets) set(config *pc.ConfigAndSecrets, mapping map[string]*pc.URL) {
	config.Secrets = mapping
}

// ProjectConfigResolver is parametric Resolver that loads values from either a
// project's configuration or secrets maps based on the type parameter.
type ProjectConfigResolver[From FromConfigOrSecrets] struct {
	Config string `help:"Load project configuration from TOML file." placeholder:"FILE" type:"existingfile"`
}

var _ Resolver = (*ProjectConfigResolver[FromConfig])(nil)

func (p ProjectConfigResolver[T]) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	mapping, err := p.getMapping(ref.Module)
	if err != nil {
		return nil, err
	}
	key, ok := mapping[ref.Name]
	if !ok {
		return nil, fmt.Errorf("no such key %q: %w", ref.Name, ErrNotFound)
	}
	return (*url.URL)(key), nil
}

func (p ProjectConfigResolver[T]) List(ctx context.Context) ([]Entry, error) {
	config, err := p.loadConfig()
	if err != nil {
		return nil, err
	}
	entries := []Entry{}
	moduleNames := maps.Keys(config.Modules)
	moduleNames = append(moduleNames, "")
	for _, moduleName := range moduleNames {
		module := optional.Zero(moduleName)
		mapping, err := p.getMapping(module)
		if err != nil {
			return nil, err
		}
		for name, key := range mapping {
			entries = append(entries, Entry{
				Ref:      Ref{module, name},
				Accessor: (*url.URL)(key),
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		im, _ := entries[i].Module.Get()
		jm, _ := entries[j].Module.Get()
		return im < jm || (im == jm && entries[i].Name < entries[j].Name)
	})
	return entries, nil
}

func (p ProjectConfigResolver[T]) Set(ctx context.Context, ref Ref, key *url.URL) error {
	mapping, err := p.getMapping(ref.Module)
	if err != nil {
		return err
	}
	mapping[ref.Name] = (*pc.URL)(key)
	return p.setMapping(ref.Module, mapping)
}

func (p ProjectConfigResolver[From]) Unset(ctx context.Context, ref Ref) error {
	mapping, err := p.getMapping(ref.Module)
	if err != nil {
		return err
	}
	delete(mapping, ref.Name)
	return p.setMapping(ref.Module, mapping)
}

func (p ProjectConfigResolver[T]) configPath() string {
	if p.Config != "" {
		return p.Config
	}
	return filepath.Join(internal.GitRoot("."), "ftl-project.toml")
}

func (p ProjectConfigResolver[T]) loadConfig() (pc.Config, error) {
	configPath := p.configPath()
	config, err := pc.Load(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return pc.Config{}, nil
	} else if err != nil {
		return pc.Config{}, err
	}
	return config, nil
}

func (p ProjectConfigResolver[T]) getMapping(module optional.Option[string]) (map[string]*pc.URL, error) {
	config, err := p.loadConfig()
	if err != nil {
		return nil, err
	}
	var t T
	if m, ok := module.Get(); ok {
		if config.Modules == nil {
			return map[string]*pc.URL{}, nil
		}
		return t.get(config.Modules[m]), nil
	}
	mapping := t.get(config.Global)
	if mapping == nil {
		mapping = map[string]*pc.URL{}
	}
	return mapping, nil
}

func (p ProjectConfigResolver[T]) setMapping(module optional.Option[string], mapping map[string]*pc.URL) error {
	config, err := p.loadConfig()
	if err != nil {
		return err
	}
	var t T
	if m, ok := module.Get(); ok {
		if config.Modules == nil {
			config.Modules = map[string]pc.ConfigAndSecrets{}
		}
		moduleConfig := config.Modules[m]
		t.set(&moduleConfig, mapping)
		config.Modules[m] = moduleConfig
	} else {
		t.set(&config.Global, mapping)
	}
	return pc.Save(p.configPath(), config)
}
