package configuration

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	pc "github.com/TBD54566975/ftl/common/projectconfig"
)

// ProjectConfigResolver is parametric Resolver that loads values from either a
// project's configuration or secrets maps based on the type parameter.
//
// See the [projectconfig] package for details on the configuration file format.
type ProjectConfigResolver[R Role] struct {
	Config []string `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`
}

var _ Resolver[Configuration] = ProjectConfigResolver[Configuration]{}
var _ Resolver[Secrets] = ProjectConfigResolver[Secrets]{}

func (p ProjectConfigResolver[R]) Role() R { var r R; return r }

func (p ProjectConfigResolver[R]) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	config, err := pc.LoadConfig(ctx, p.Config)
	if err != nil {
		return nil, err
	}
	mapping, err := p.getMapping(config, ref.Module)
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
	config, err := pc.LoadConfig(ctx, p.Config)
	if err != nil {
		return nil, err
	}
	entries := []Entry{}
	moduleNames := maps.Keys(config.Modules)
	moduleNames = append(moduleNames, "")
	for _, moduleName := range moduleNames {
		module := optional.Zero(moduleName)
		mapping, err := p.getMapping(config, module)
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
	config, err := pc.LoadWritableConfig(ctx, p.Config)
	if err != nil {
		return err
	}
	mapping, err := p.getMapping(config, ref.Module)
	if err != nil {
		return err
	}
	mapping[ref.Name] = (*pc.URL)(key)
	return p.setMapping(config, ref.Module, mapping)
}

func (p ProjectConfigResolver[From]) Unset(ctx context.Context, ref Ref) error {
	config, err := pc.LoadWritableConfig(ctx, p.Config)
	if err != nil {
		return err
	}
	mapping, err := p.getMapping(config, ref.Module)
	if err != nil {
		return err
	}
	delete(mapping, ref.Name)
	return p.setMapping(config, ref.Module, mapping)
}

func (p ProjectConfigResolver[R]) getMapping(config pc.Config, module optional.Option[string]) (map[string]*pc.URL, error) {
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
	return mapping, nil
}

func (p ProjectConfigResolver[R]) setMapping(config pc.Config, module optional.Option[string], mapping map[string]*pc.URL) error {
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
	configPaths := pc.ConfigPaths(p.Config)
	return pc.Save(configPaths[len(configPaths)-1], config)
}
