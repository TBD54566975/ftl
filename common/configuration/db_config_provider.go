package configuration

import (
	"context"
	"errors"
	"net/url"

	"github.com/alecthomas/types/optional"

	dalerrs "github.com/TBD54566975/ftl/backend/dal"
)

// DBProvider is a configuration or secrets provider that stores data in its key.
type DBProvider[R Role] struct {
	dal DBProviderDAL
}

type DBProviderDAL interface {
	GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error)
	SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error
	UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error

	GetModuleSecret(ctx context.Context, module optional.Option[string], name string) ([]byte, error)
	SetModuleSecret(ctx context.Context, module optional.Option[string], name string, value []byte) error
	UnsetModuleSecret(ctx context.Context, module optional.Option[string], name string) error
}

func NewDBProvider[R Role](dal DBProviderDAL) DBProvider[R] {
	return DBProvider[R]{
		dal: dal,
	}
}

func (DBProvider[R]) Role() R     { var r R; return r }
func (DBProvider[R]) Key() string { return "db" }

func (d DBProvider[R]) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	var value []byte
	var err error

	switch any(new(R)).(type) {
	case *Configuration:
		value, err = d.dal.GetModuleConfiguration(ctx, ref.Module, ref.Name)
	case *Secrets:
		value, err = d.dal.GetModuleSecret(ctx, ref.Module, ref.Name)
	}

	if err != nil {
		return nil, dalerrs.ErrNotFound
	}
	return value, nil
}

func (d DBProvider[R]) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	var err error

	switch any(new(R)).(type) {
	case *Configuration:
		err = d.dal.SetModuleConfiguration(ctx, ref.Module, ref.Name, value)
	case *Secrets:
		err = d.dal.SetModuleSecret(ctx, ref.Module, ref.Name, value)
	}

	if err != nil {
		return nil, err
	}
	return &url.URL{Scheme: "db"}, nil
}

func (d DBProvider[R]) Delete(ctx context.Context, ref Ref) error {
	switch any(new(R)).(type) {
	case *Configuration:
		return d.dal.UnsetModuleConfiguration(ctx, ref.Module, ref.Name)
	case *Secrets:
		return d.dal.UnsetModuleSecret(ctx, ref.Module, ref.Name)
	default:
		return errors.New("unknown role")
	}
}
