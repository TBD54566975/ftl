package configuration

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	Config []string `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`
}

var _ Resolver[Configuration] = ProjectConfigResolver[Configuration]{}
var _ Resolver[Secrets] = ProjectConfigResolver[Secrets]{}

func (p ProjectConfigResolver[R]) Role() R { var r R; return r }

func (p ProjectConfigResolver[R]) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	config, err := p.LoadConfig(ctx)
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
	config, err := p.LoadConfig(ctx)
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
	config, err := p.loadWritableConfig(ctx)
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
	config, err := p.loadWritableConfig(ctx)
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

// ConfigPaths returns the computed list of configuration paths to load.
func (p ProjectConfigResolver[R]) ConfigPaths() []string {
	if len(p.Config) > 0 {
		return p.Config
	}
	path := filepath.Join(internal.GitRoot(""), "ftl-project.toml")
	_, err := os.Stat(path)
	if err == nil {
		return []string{path}
	}
	return []string{}
}

func (p ProjectConfigResolver[R]) loadWritableConfig(ctx context.Context) (pc.Config, error) {
	configPaths := p.ConfigPaths()
	if len(configPaths) == 0 {
		return pc.Config{}, nil
	}
	target := configPaths[len(configPaths)-1]
	log.FromContext(ctx).Tracef("Loading config from %s", target)
	return pc.Load(target)
}

func (p ProjectConfigResolver[R]) LoadConfig(ctx context.Context) (pc.Config, error) {
	logger := log.FromContext(ctx)
	configPaths := p.ConfigPaths()
	logger.Tracef("Loading config from %s", strings.Join(configPaths, " "))
	config, err := pc.Merge(configPaths...)
	if err != nil {
		return pc.Config{}, err
	}
	return config, nil
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
	configPaths := p.ConfigPaths()
	return pc.Save(configPaths[len(configPaths)-1], config)
}
