package server

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/alecthomas/types/once"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/modulecontext"
)

func PostgresDatabaseHandle[T ftl.DatabaseConfig]() reflection.VerbResource {
	typ := reflect.TypeFor[T]()
	var config T
	if typ.Kind() == reflect.Ptr {
		config = reflect.New(typ.Elem()).Interface().(T) //nolint:forcetypeassert
	} else {
		config = reflect.New(typ).Elem().Interface().(T) //nolint:forcetypeassert
	}

	return func() reflect.Value {
		reflectedDB := reflection.GetDatabase[T]()
		db := ftl.NewDatabaseHandle(config, ftl.DatabaseTypePostgres, reflectedDB.DB)
		return reflect.ValueOf(db)
	}
}

func InitPostgres(ref reflection.Ref) *reflection.ReflectedDatabaseHandle {
	return &reflection.ReflectedDatabaseHandle{
		Name:   ref.Name,
		DBType: "postgres",
		DB: once.Once(func(ctx context.Context) (*sql.DB, error) {
			logger := log.FromContext(ctx)

			provider := modulecontext.FromContext(ctx).CurrentContext()
			dsn, err := provider.GetDatabase(ref.Name, modulecontext.DBTypePostgres)
			if err != nil {
				return nil, fmt.Errorf("failed to get database %q: %w", ref.Name, err)
			}

			logger.Debugf("Opening database: %s", ref.Name)
			db, err := otelsql.Open("pgx", dsn)
			if err != nil {
				return nil, fmt.Errorf("failed to open database %q: %w", ref.Name, err)
			}

			// sets db.system and db.name attributes
			metricAttrs := otelsql.WithAttributes(
				semconv.DBSystemPostgreSQL,
				semconv.DBNameKey.String(ref.Name),
				attribute.Bool("ftl.is_user_service", true),
			)
			err = otelsql.RegisterDBStatsMetrics(db, metricAttrs)
			if err != nil {
				return nil, fmt.Errorf("failed to register database metrics: %w", err)
			}
			db.SetConnMaxIdleTime(time.Minute)
			db.SetMaxOpenConns(20)
			return db, nil
		}),
	}
}
