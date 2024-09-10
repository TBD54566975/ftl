package profiles

import (
	"context"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/configuration/routers"
	"github.com/TBD54566975/ftl/internal/profiles/internal"
)

type ProjectConfig internal.Project

type Config struct {
	Name     string
	Endpoint *url.URL
}

type Profile struct {
	shared ProjectConfig
	config Config
	sm     *manager.Manager[configuration.Secrets]
	cm     *manager.Manager[configuration.Configuration]
}

// ProjectConfig is static project-wide configuration shared by all profiles.
func (p *Profile) ProjectConfig() ProjectConfig { return p.shared }

// Config is the static configuration for a Profile.
func (p *Profile) Config() Config { return p.config }

func (p *Profile) SecretsManager() *manager.Manager[configuration.Secrets]             { return p.sm }
func (p *Profile) ConfigurationManager() *manager.Manager[configuration.Configuration] { return p.cm }

// Init a new project with a default "local" profile.
func Init(project ProjectConfig) error {
	err := internal.Init(internal.Project(project))
	if err != nil {
		return fmt.Errorf("init project: %w", err)
	}
	return nil
}

// Load a profile from the project.
func Load(
	ctx context.Context,
	secretsRegistry *providers.Registry[configuration.Secrets],
	configRegistry *providers.Registry[configuration.Configuration],
	root string,
	profile string,
) (Profile, error) {
	project, err := internal.Load(root)
	if err != nil {
		return Profile{}, fmt.Errorf("load project: %w", err)
	}
	prof, err := project.LoadProfile(profile)
	if err != nil {
		return Profile{}, fmt.Errorf("load profile: %w", err)
	}
	profileEndpoint, err := prof.EndpointURL()
	if err != nil {
		return Profile{}, fmt.Errorf("profile endpoint: %w", err)
	}

	var sm *manager.Manager[configuration.Secrets]
	var cm *manager.Manager[configuration.Configuration]
	switch prof.Type {
	case internal.ProfileTypeLocal:
		sp, err := secretsRegistry.Get(ctx, prof.SecretsProvider)
		if err != nil {
			return Profile{}, fmt.Errorf("get secrets provider: %w", err)
		}
		secretsRouter := routers.NewFileRouter[configuration.Secrets](project.LocalSecretsPath(profile))
		sm, err = manager.New[configuration.Secrets](ctx, secretsRouter, []configuration.Provider[configuration.Secrets]{sp})
		if err != nil {
			return Profile{}, fmt.Errorf("create secrets manager: %w", err)
		}

		cp, err := configRegistry.Get(ctx, prof.ConfigProvider)
		if err != nil {
			return Profile{}, fmt.Errorf("get config provider: %w", err)
		}
		configRouter := routers.NewFileRouter[configuration.Configuration](project.LocalConfigPath(profile))
		cm, err = manager.New[configuration.Configuration](ctx, configRouter, []configuration.Provider[configuration.Configuration]{cp})
		if err != nil {
			return Profile{}, fmt.Errorf("create configuration manager: %w", err)
		}

	case internal.ProfileTypeRemote:
		panic("not implemented")

	default:
		return Profile{}, fmt.Errorf("%s: unknown profile type: %q", profile, prof.Type)
	}
	return Profile{
		shared: ProjectConfig(project),
		config: Config{
			Name:     prof.Name,
			Endpoint: profileEndpoint,
		},
		sm: sm,
		cm: cm,
	}, nil
}
