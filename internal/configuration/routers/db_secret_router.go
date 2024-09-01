package routers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/dal"
)

// DatabaseSecrets loads values a project's secrets from the given database.
type DatabaseSecrets struct {
	dal DatabaseSecretsDAL
}

type DatabaseSecretsDAL interface {
	GetModuleSecretURL(ctx context.Context, module optional.Option[string], name string) (string, error)
	ListModuleSecrets(ctx context.Context) ([]dal.ModuleSecret, error)
	SetModuleSecretURL(ctx context.Context, module optional.Option[string], name string, url string) error
	UnsetModuleSecret(ctx context.Context, module optional.Option[string], name string) error
}

// DatabaseSecrets should only be used for secrets
var _ configuration.Router[configuration.Secrets] = DatabaseSecrets{}

func NewDatabaseSecrets(db DatabaseSecretsDAL) DatabaseSecrets {
	return DatabaseSecrets{dal: db}
}

func (d DatabaseSecrets) Role() configuration.Secrets { return configuration.Secrets{} }

func (d DatabaseSecrets) Get(ctx context.Context, ref configuration.Ref) (*url.URL, error) {
	u, err := d.dal.GetModuleSecretURL(ctx, ref.Module, ref.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret URL: %w", err)
	}
	url, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secret URL: %w", err)
	}
	return url, nil
}

func (d DatabaseSecrets) List(ctx context.Context) ([]configuration.Entry, error) {
	secrets, err := d.dal.ListModuleSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list module secrets: %w", err)
	}
	entries := make([]configuration.Entry, len(secrets))
	for i, s := range secrets {
		url, err := url.Parse(s.Url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse secret URL: %w", err)
		}
		entries[i] = configuration.Entry{
			Ref: configuration.Ref{
				Module: s.Module,
				Name:   s.Name,
			},
			Accessor: url,
		}
	}
	return entries, nil
}

func (d DatabaseSecrets) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	err := d.dal.SetModuleSecretURL(ctx, ref.Module, ref.Name, key.String())
	if err != nil {
		return fmt.Errorf("failed to set secret URL: %w", err)
	}
	return nil
}

func (d DatabaseSecrets) Unset(ctx context.Context, ref configuration.Ref) error {
	err := d.dal.UnsetModuleSecret(ctx, ref.Module, ref.Name)
	if err != nil {
		return fmt.Errorf("failed to unset secret: %w", err)
	}
	return nil
}
