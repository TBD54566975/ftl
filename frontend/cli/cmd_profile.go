package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alecthomas/types/either"

	"github.com/block/ftl"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/providers"
	"github.com/block/ftl/internal/profiles"
)

type profileCmd struct {
	Init    profileInitCmd    `cmd:"" help:"Initialize a new project."`
	List    profileListCmd    `cmd:"" help:"List all profiles."`
	Default profileDefaultCmd `cmd:"" help:"Set a profile as default."`
	Switch  profileSwitchCmd  `cmd:"" help:"Switch locally active profile."`
	New     profileNewCmd     `cmd:"" help:"Create a new local or remote profile."`
}

type profileInitCmd struct {
	Project     string   `arg:"" help:"Name of the project."`
	Dir         string   `arg:"" help:"Directory to initialize the project in." default:"${gitroot}" required:""`
	ModuleRoots []string `help:"Root directories of existing modules."`
	NoGit       bool     `help:"Don't add files to the git repository."`
}

func (p profileInitCmd) Run(
	configRegistry *providers.Registry[configuration.Configuration],
	secretsRegistry *providers.Registry[configuration.Secrets],
) error {
	_, err := profiles.Init(profiles.ProjectConfig{
		Realm:         p.Project,
		FTLMinVersion: ftl.Version,
		ModuleRoots:   p.ModuleRoots,
		NoGit:         p.NoGit,
		Root:          p.Dir,
	}, secretsRegistry, configRegistry)
	if err != nil {
		return fmt.Errorf("init project: %w", err)
	}
	fmt.Printf("Project initialized in %s.\n", p.Dir)
	return nil
}

type profileListCmd struct{}

func (profileListCmd) Run(project *profiles.Project) error {
	active, err := project.ActiveProfile()
	if err != nil {
		return fmt.Errorf("active profile: %w", err)
	}
	p, err := project.List()
	if err != nil {
		return fmt.Errorf("list profiles: %w", err)
	}
	for _, profile := range p {
		attrs := []string{}
		switch profile.Config.(type) {
		case either.Left[profiles.LocalProfileConfig, profiles.RemoteProfileConfig]:
			attrs = append(attrs, "local")
		case either.Right[profiles.LocalProfileConfig, profiles.RemoteProfileConfig]:
			attrs = append(attrs, "remote")
		}
		if project.DefaultProfile() == profile.Name {
			attrs = append(attrs, "default")
		}
		if active == profile.Name {
			attrs = append(attrs, "active")
		}
		fmt.Printf("%s (%s)\n", profile, strings.Join(attrs, ", "))
	}
	return nil
}

type profileDefaultCmd struct {
	Profile string `arg:"" help:"Profile name."`
}

func (p profileDefaultCmd) Run(project *profiles.Project) error {
	err := project.SetDefault(p.Profile)
	if err != nil {
		return fmt.Errorf("set default profile: %w", err)
	}
	return nil
}

type profileSwitchCmd struct {
	Profile string `arg:"" help:"Profile name."`
}

func (p profileSwitchCmd) Run(project *profiles.Project) error {
	err := project.Switch(p.Profile)
	if err != nil {
		return fmt.Errorf("switch profile: %w", err)
	}
	return nil
}

type profileNewCmd struct {
	Local         bool                      `help:"Create a local profile." xor:"location" and:"providers"`
	Remote        *url.URL                  `help:"Create a remote profile." xor:"location" placeholder:"ENDPOINT"`
	Secrets       configuration.ProviderKey `help:"Secrets provider." default:"inline" and:"providers"`
	Configuration configuration.ProviderKey `help:"Configuration provider." default:"inline" and:"providers"`
	Name          string                    `arg:"" help:"Profile name."`
}

func (profileNewCmd) Help() string {
	return `
Specify either --local or --remote=ENDPOINT to create a new profile.

A local profile (specified via --local) is used for local development and testing, and can be managed without a running
FTL cluster. In a local profile, secrets and configuration are stored in locally accessible secret stores, including
1Password (--secrets=op), Keychain (--secrets=keychain), and local files (--secrets=inline).

A remote profile (specified via --remote=ENDPOINT) is used for persistent cloud deployments. In a remote profile, secrets
and configuration are managed by the FTL cluster.

eg.

Create a new local profile with secrets stored in the Keychain, and configuration stored inline:

    ftl profile new devel --local --secrets=keychain

Create a new remote profile:

    ftl profile new staging --remote=https://ftl.example.com
`
}

func (p profileNewCmd) Run(project *profiles.Project) error {
	var config either.Either[profiles.LocalProfileConfig, profiles.RemoteProfileConfig]
	switch {
	case p.Local:
		config = either.LeftOf[profiles.RemoteProfileConfig](profiles.LocalProfileConfig{
			SecretsProvider: p.Secrets,
			ConfigProvider:  p.Configuration,
		})

	case p.Remote != nil:
		config = either.RightOf[profiles.LocalProfileConfig](profiles.RemoteProfileConfig{
			Endpoint: p.Remote,
		})
	}
	err := project.New(profiles.ProfileConfig{
		Name:   p.Name,
		Config: config,
	})
	if err != nil {
		return fmt.Errorf("new profile: %w", err)
	}
	return nil
}
