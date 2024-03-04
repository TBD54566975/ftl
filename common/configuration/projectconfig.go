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
	"github.com/TBD54566975/ftl/internal/log"
)

// ProjectConfigResolver is parametric Resolver that loads values from either a
// project's configuration or secrets maps based on the type parameter.
//
// See the [projectconfig] package for details on the configuration file format.
type ProjectConfigResolver[R Role] struct {
	Config string `help:"Load project configuration file." placeholder:"FILE" type:"existingfile" env:"FTL_CONFIG"`
}

var _ Resolver = ProjectConfigResolver[Configuration]{}
var _ Resolver = ProjectConfigResolver[Secrets]{}

func (p ProjectConfigResolver[R]) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	mapping, err := p.getMapping(ctx, ref.Module)
	if err != nil {
		return nil, err
	}
	key, ok := mapping[ref.Name]
	if !ok {
		return nil, fmt.Errorf("no such key %q: %w", ref.Name, ErrNotFound)
	}
	return (*url.URL)(key), nil
}

func (p ProjectConfigResolver[R]) List(ctx context.Context) ([]Entry, error) {
	config, err := p.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	entries := []Entry{}
	moduleNames := maps.Keys(config.Modules)
	moduleNames = append(moduleNames, "")
	for _, moduleName := range moduleNames {
		module := optional.Zero(moduleName)
		mapping, err := p.getMapping(ctx, module)
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
	sort.SliceStable(entries, func(i, j int) bool {
		im, _ := entries[i].Module.Get()
		jm, _ := entries[j].Module.Get()
		return im < jm || (im == jm && entries[i].Name < entries[j].Name)
	})
	return entries, nil
}

func (p ProjectConfigResolver[R]) Set(ctx context.Context, ref Ref, key *url.URL) error {
	mapping, err := p.getMapping(ctx, ref.Module)
	if err != nil {
		return err
	}
	mapping[ref.Name] = (*pc.URL)(key)
	return p.setMapping(ctx, ref.Module, mapping)
}

func (p ProjectConfigResolver[From]) Unset(ctx context.Context, ref Ref) error {
	mapping, err := p.getMapping(ctx, ref.Module)
	if err != nil {
		return err
	}
	delete(mapping, ref.Name)
	return p.setMapping(ctx, ref.Module, mapping)
}

func (p ProjectConfigResolver[R]) configPath() string {
	if p.Config != "" {
		return p.Config
	}
	return filepath.Join(internal.GitRoot(""), "ftl-project.toml")
}

func (p ProjectConfigResolver[R]) loadConfig(ctx context.Context) (pc.Config, error) {
	logger := log.FromContext(ctx)
	configPath := p.configPath()
	logger.Tracef("Loading config from %s", configPath)
	config, err := pc.Load(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return pc.Config{}, nil
	} else if err != nil {
		return pc.Config{}, err
	}
	return config, nil
}

func (p ProjectConfigResolver[R]) getMapping(ctx context.Context, module optional.Option[string]) (map[string]*pc.URL, error) {
	config, err := p.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	var k R
	get := func(dest pc.ConfigAndSecrets) map[string]*pc.URL {
		switch any(k).(type) {
		case Configuration:
			return dest.Config
		case Secrets:
			return dest.Secrets
		default:
			panic("unsupported kind")
		}
	}

	var mapping map[string]*pc.URL
	if m, ok := module.Get(); ok {
		if config.Modules == nil {
			return map[string]*pc.URL{}, nil
		}
		mapping = get(config.Modules[m])
	} else {
		mapping = get(config.Global)
	}
	if mapping == nil {
		return map[string]*pc.URL{}, nil
	}
	return mapping, nil
}

func (p ProjectConfigResolver[R]) setMapping(ctx context.Context, module optional.Option[string], mapping map[string]*pc.URL) error {
	config, err := p.loadConfig(ctx)
	if err != nil {
		return err
	}

	var k R
	set := func(dest *pc.ConfigAndSecrets, mapping map[string]*pc.URL) {
		switch any(k).(type) {
		case Configuration:
			dest.Config = mapping
		case Secrets:
			dest.Secrets = mapping
		}
	}

	if m, ok := module.Get(); ok {
		if config.Modules == nil {
			config.Modules = map[string]pc.ConfigAndSecrets{}
		}
		moduleConfig := config.Modules[m]
		set(&moduleConfig, mapping)
		config.Modules[m] = moduleConfig
	} else {
		set(&config.Global, mapping)
	}
	return pc.Save(p.configPath(), config)
}
