package routers

import (
	"context"
	"net/url"
	"sort"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/internal/configuration"
	pc "github.com/block/ftl/internal/projectconfig"
)

// ProjectConfig is parametric Resolver that loads values from either a
// project's configuration or secrets maps based on the type parameter.
//
// See the [projectconfig] package for details on the configuration file format.
type ProjectConfig[R configuration.Role] struct {
	Config string `name:"config" short:"C" help:"Path to FTL project configuration file." env:"FTL_CONFIG" placeholder:"FILE" type:"existingfile"`
}

var _ configuration.Router[configuration.Configuration] = ProjectConfig[configuration.Configuration]{}
var _ configuration.Router[configuration.Secrets] = ProjectConfig[configuration.Secrets]{}

func (p ProjectConfig[R]) Role() R { var r R; return r }

func (p ProjectConfig[R]) Get(ctx context.Context, ref configuration.Ref) (*url.URL, error) {
	config, err := pc.Load(ctx, p.Config)
	if err != nil {
		return nil, err
	}
	mapping, err := p.getMapping(config, ref.Module)
	if err != nil {
		return nil, err
	}
	key, ok := mapping[ref.Name]
	if !ok {
		return nil, configuration.ErrNotFound
	}
	return (*url.URL)(key), nil
}

func (p ProjectConfig[R]) List(ctx context.Context) ([]configuration.Entry, error) {
	config, err := pc.Load(ctx, p.Config)
	if err != nil {
		return nil, err
	}
	entries := []configuration.Entry{}
	moduleNames := maps.Keys(config.Modules)
	moduleNames = append(moduleNames, "")
	for _, moduleName := range moduleNames {
		module := optional.Zero(moduleName)
		mapping, err := p.getMapping(config, module)
		if err != nil {
			return nil, err
		}
		for name, key := range mapping {
			entries = append(entries, configuration.Entry{
				Ref:      configuration.Ref{module, name},
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

func (p ProjectConfig[R]) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	config, err := pc.Load(ctx, p.Config)
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

func (p ProjectConfig[From]) Unset(ctx context.Context, ref configuration.Ref) error {
	config, err := pc.Load(ctx, p.Config)
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

func (p ProjectConfig[R]) getMapping(config pc.Config, module optional.Option[string]) (map[string]*pc.URL, error) {
	var k R
	get := func(dest pc.ConfigAndSecrets) map[string]*pc.URL {
		switch any(k).(type) {
		case configuration.Configuration:
			return emptyMapIfNil(dest.Config)
		case configuration.Secrets:
			return emptyMapIfNil(dest.Secrets)
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

func emptyMapIfNil(mapping map[string]*pc.URL) map[string]*pc.URL {
	if mapping == nil {
		return map[string]*pc.URL{}
	}
	return mapping
}

func (p ProjectConfig[R]) setMapping(config pc.Config, module optional.Option[string], mapping map[string]*pc.URL) error {
	var k R
	set := func(dest *pc.ConfigAndSecrets, mapping map[string]*pc.URL) {
		switch any(k).(type) {
		case configuration.Configuration:
			dest.Config = mapping
		case configuration.Secrets:
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
	return pc.Save(config)
}
