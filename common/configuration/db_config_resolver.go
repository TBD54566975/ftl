package configuration

import (
	"context"
	"errors"
	"net/url"

	"github.com/TBD54566975/ftl/common/configuration/sql"
	"github.com/TBD54566975/ftl/internal/slices"
)

// DBResolver loads values a project's configuration from the given database.
type DBResolver[R Role] struct {
	dal DBResolverDAL
}

type DBResolverDAL interface {
	ListModuleConfiguration(ctx context.Context) ([]sql.ModuleConfiguration, error)
	ListModuleSecrets(ctx context.Context) ([]sql.ModuleSecret, error)
}

var _ Router[Configuration] = DBResolver[Configuration]{}
var _ Router[Secrets] = DBResolver[Secrets]{}

func NewDBResolver[R Role](db DBResolverDAL) DBResolver[R] {
	return DBResolver[R]{dal: db}
}

func (DBResolver[R]) Role() R { var r R; return r }

func (d DBResolver[R]) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	return urlPtr(), nil
}

func (d DBResolver[R]) List(ctx context.Context) ([]Entry, error) {
	switch any(new(R)).(type) {
	case *Configuration:
		configs, err := d.dal.ListModuleConfiguration(ctx)
		if err != nil {
			return nil, err
		}
		return slices.Map(configs, func(c sql.ModuleConfiguration) Entry {
			return Entry{
				Ref: Ref{
					Module: c.Module,
					Name:   c.Name,
				},
				Accessor: urlPtr(),
			}
		}), nil
	case *Secrets:
		secrets, err := d.dal.ListModuleSecrets(ctx)
		if err != nil {
			return nil, err
		}
		return slices.Map(secrets, func(s sql.ModuleSecret) Entry {
			return Entry{
				Ref: Ref{
					Module: s.Module,
					Name:   s.Name,
				},
				Accessor: urlPtr(),
			}
		}), nil
	default:
		return nil, errors.New("unknown role")
	}
}

func (d DBResolver[R]) Set(ctx context.Context, ref Ref, key *url.URL) error {
	// Writing to the DB is performed by DBProvider, so this function is a NOOP
	return nil
}

func (d DBResolver[R]) Unset(ctx context.Context, ref Ref) error {
	// Writing to the DB is performed by DBProvider, so this function is a NOOP
	return nil
}

func (d DBResolver[R]) UseWithProvider(ctx context.Context, pkey string) bool {
	return pkey == "db"
}

func urlPtr() *url.URL {
	// The URLs for Database-provided configs are not actually used because all the
	// information needed to load each config is contained in the Ref, so we pass
	// around an empty "db://" to satisfy the expectations of the Resolver interface.
	return &url.URL{Scheme: "db"}
}
