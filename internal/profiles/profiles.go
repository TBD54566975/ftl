package profiles

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/alecthomas/types/either"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/configuration/routers"
	"github.com/TBD54566975/ftl/internal/profiles/internal"
	"github.com/TBD54566975/ftl/internal/reflect"
	"github.com/TBD54566975/ftl/internal/slices"
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

// ProjectConfig is the static project-wide configuration shared by all profiles.
func (p *Profile) ProjectConfig() ProjectConfig { return p.shared }

// Config is the static configuration for a Profile.
func (p *Profile) Config() Config { return reflect.DeepCopy(p.config) }

// SecretsManager returns the secrets manager for this profile.
func (p *Profile) SecretsManager() *manager.Manager[configuration.Secrets] { return p.sm }

// ConfigurationManager returns the configuration manager for this profile.
func (p *Profile) ConfigurationManager() *manager.Manager[configuration.Configuration] { return p.cm }

type LocalProfileConfig struct {
	SecretsProvider configuration.ProviderKey
	ConfigProvider  configuration.ProviderKey
}

type RemoteProfileConfig struct {
	Endpoint *url.URL
}

type ProfileConfig struct {
	Name   string
	Config either.Either[LocalProfileConfig, RemoteProfileConfig]
}

func (p ProfileConfig) String() string { return p.Name }

type Project struct {
	project         internal.Project
	secretsRegistry *providers.Registry[configuration.Secrets]
	configRegistry  *providers.Registry[configuration.Configuration]
}

// Open a project.
func Open(
	root string,
	secretsRegistry *providers.Registry[configuration.Secrets],
	configRegistry *providers.Registry[configuration.Configuration],
) (*Project, error) {
	project, err := internal.Load(root)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	return &Project{
		project:         project,
		secretsRegistry: secretsRegistry,
		configRegistry:  configRegistry,
	}, nil
}

// Init a new project with a default local profile.
//
// If "project.Default" is empty a new project will be created with a default "local" profile.
func Init(
	project ProjectConfig,
	secretsRegistry *providers.Registry[configuration.Secrets],
	configRegistry *providers.Registry[configuration.Configuration],
) (*Project, error) {
	err := internal.Init(internal.Project(project))
	if err != nil {
		return nil, fmt.Errorf("init project: %w", err)
	}
	return &Project{
		project:         internal.Project(project),
		secretsRegistry: secretsRegistry,
		configRegistry:  configRegistry,
	}, nil
}

// SetDefault profile for the project.
func (p *Project) SetDefault(profile string) error {
	_, err := p.project.LoadProfile(profile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s: profile does not exist", profile)
		}
		return fmt.Errorf("%s: load profile: %w", profile, err)
	}
	p.project.DefaultProfile = profile
	err = p.project.Save()
	if err != nil {
		return fmt.Errorf("%s: save project: %w", profile, err)
	}
	return nil
}

// Switch active profiles.
func (p *Project) Switch(profile string) error {
	_, err := p.project.LoadProfile(profile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s: profile does not exist", profile)
		}
		return fmt.Errorf("%s: load profile: %w", profile, err)
	}
	err = p.project.SetActiveProfile(profile)
	if err != nil {
		return fmt.Errorf("set active profile: %w", err)
	}
	return nil
}

// ActiveProfile returns the name of the active profile.
//
// If no profile is active, the default profile is returned.
func (p *Project) ActiveProfile() (string, error) {
	profile, err := p.project.ActiveProfile()
	if err != nil {
		return "", fmt.Errorf("active profile: %w", err)
	}
	return profile, nil
}

// DefaultProfile returns the name of the default profile.
func (p *Project) DefaultProfile() string {
	return p.project.DefaultProfile
}

// List all profiles in the project.
func (p *Project) List() ([]ProfileConfig, error) {
	profiles, err := p.project.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("load profiles: %w", err)
	}
	configs, err := slices.MapErr(profiles, func(profile internal.Profile) (ProfileConfig, error) {
		var config either.Either[LocalProfileConfig, RemoteProfileConfig]
		switch profile.Type {
		case internal.ProfileTypeLocal:
			config = either.LeftOf[RemoteProfileConfig](LocalProfileConfig{
				SecretsProvider: profile.SecretsProvider,
				ConfigProvider:  profile.ConfigProvider,
			})
		case internal.ProfileTypeRemote:
			endpoint, err := profile.EndpointURL()
			if err != nil {
				return ProfileConfig{}, fmt.Errorf("profile endpoint: %w", err)
			}
			config = either.RightOf[LocalProfileConfig](RemoteProfileConfig{
				Endpoint: endpoint,
			})
		}
		return ProfileConfig{
			Name:   profile.Name,
			Config: config,
		}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("map profiles: %w", err)
	}
	return configs, nil
}

// New creates a new profile in the project.
func (p *Project) New(profileConfig ProfileConfig) error {
	_, err := p.project.LoadProfile(profileConfig.Name)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("load profile: %w", err)
		}
	} else {
		return fmt.Errorf("profile %s already exists", profileConfig.Name)
	}
	var profile internal.Profile
	switch handle := profileConfig.Config.(type) {
	case either.Left[LocalProfileConfig, RemoteProfileConfig]:
		config := handle.Get()
		profile = internal.Profile{
			Name:            profileConfig.Name,
			Type:            internal.ProfileTypeLocal,
			SecretsProvider: config.SecretsProvider,
			ConfigProvider:  config.ConfigProvider,
		}

	case either.Right[LocalProfileConfig, RemoteProfileConfig]:
		config := handle.Get()
		profile = internal.Profile{
			Name:     profileConfig.Name,
			Endpoint: config.Endpoint.String(),
			Type:     internal.ProfileTypeRemote,
		}

	case nil:
		return fmt.Errorf("profile config is nil")
	}
	err = p.project.SaveProfile(profile)
	if err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	return nil
}

// Load a profile from the project.
func (p *Project) Load(ctx context.Context, profile string) (Profile, error) {
	prof, err := p.project.LoadProfile(profile)
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
		sp, err := p.secretsRegistry.Get(ctx, prof.SecretsProvider)
		if err != nil {
			return Profile{}, fmt.Errorf("get secrets provider: %w", err)
		}
		secretsRouter := routers.NewFileRouter[configuration.Secrets](p.project.LocalSecretsPath(profile))
		sm, err = manager.New[configuration.Secrets](ctx, secretsRouter, []configuration.Provider[configuration.Secrets]{sp})
		if err != nil {
			return Profile{}, fmt.Errorf("create secrets manager: %w", err)
		}

		cp, err := p.configRegistry.Get(ctx, prof.ConfigProvider)
		if err != nil {
			return Profile{}, fmt.Errorf("get config provider: %w", err)
		}
		configRouter := routers.NewFileRouter[configuration.Configuration](p.project.LocalConfigPath(profile))
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
		shared: ProjectConfig(p.project),
		config: Config{
			Name:     prof.Name,
			Endpoint: profileEndpoint,
		},
		sm: sm,
		cm: cm,
	}, nil
}
