package configuration

import (
	"context"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/common/configuration/sql"
	"github.com/alecthomas/types/optional"
)

// DBSecretResolver loads values a project's secrets from the given database.
type DBSecretResolver struct {
	dal DBSecretResolverDAL
}

type DBSecretResolverDAL interface {
	GetModuleSecretURL(ctx context.Context, module optional.Option[string], name string) (string, error)
	ListModuleSecrets(ctx context.Context) ([]sql.ModuleSecret, error)
	SetModuleSecretURL(ctx context.Context, module optional.Option[string], name string, url string) error
	UnsetModuleSecret(ctx context.Context, module optional.Option[string], name string) error
}

// DBSecretResolver should only be used for secrets
var _ Router[Secrets] = DBSecretResolver{}

func NewDBSecretResolver(db DBSecretResolverDAL) DBSecretResolver {
	return DBSecretResolver{dal: db}
}

func (d DBSecretResolver) Role() Secrets { return Secrets{} }

func (d DBSecretResolver) Get(ctx context.Context, ref Ref) (*url.URL, error) {
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

func (d DBSecretResolver) List(ctx context.Context) ([]Entry, error) {
	secrets, err := d.dal.ListModuleSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list module secrets: %w", err)
	}
	entries := make([]Entry, len(secrets))
	for i, s := range secrets {
		url, err := url.Parse(s.Url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse secret URL: %w", err)
		}
		entries[i] = Entry{
			Ref: Ref{
				Module: s.Module,
				Name:   s.Name,
			},
			Accessor: url,
		}
	}
	return entries, nil
}

func (d DBSecretResolver) Set(ctx context.Context, ref Ref, key *url.URL) error {
	err := d.dal.SetModuleSecretURL(ctx, ref.Module, ref.Name, key.String())
	if err != nil {
		return fmt.Errorf("failed to set secret URL: %w", err)
	}
	return nil
}

func (d DBSecretResolver) Unset(ctx context.Context, ref Ref) error {
	err := d.dal.UnsetModuleSecret(ctx, ref.Module, ref.Name)
	if err != nil {
		return fmt.Errorf("failed to unset secret: %w", err)
	}
	return nil
}
