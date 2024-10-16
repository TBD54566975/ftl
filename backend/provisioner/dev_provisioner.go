package provisioner

import (
	"context"
	"fmt"

	"github.com/XSAM/otelsql"

	"github.com/TBD54566975/ftl/backend/controller/dsn"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/internal/dev"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

// NewDevProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode
func NewDevProvisioner(postgresPort int) *InMemProvisioner {
	var postgresDSN string
	return NewEmbeddedProvisioner(map[ResourceType]InMemResourceProvisionerFn{
		ResourceTypePostgres: func(ctx context.Context, rc *provisioner.ResourceContext, module, id string) (*provisioner.Resource, error) {
			pg, ok := rc.Resource.Resource.(*provisioner.Resource_Postgres)
			if !ok {
				panic(fmt.Errorf("unexpected resource type: %T", rc.Resource.Resource))
			}
			logger := log.FromContext(ctx)
			logger.Infof("provisioning postgres database: %s_%s", module, id)

			dbName := strcase.ToLowerCamel(module) + "_" + strcase.ToLowerCamel(id)

			if postgresDSN == "" {
				// We assume that the DB hsas already been started when running in dev mode
				pdsn, err := dev.WaitForDBReady(ctx, postgresPort)
				if err != nil {
					return nil, fmt.Errorf("failed to wait for postgres to be ready: %w", err)
				}
				postgresDSN = pdsn
			}
			conn, err := otelsql.Open("pgx", postgresDSN)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to postgres: %w", err)
			}
			defer conn.Close()

			res, err := conn.Query("SELECT * FROM pg_catalog.pg_database WHERE datname=$1", dbName)
			if err != nil {
				return nil, fmt.Errorf("failed to query database: %w", err)
			}
			defer res.Close()
			if !res.Next() {
				_, err = conn.ExecContext(ctx, "CREATE DATABASE "+dbName)
				if err != nil {
					return nil, fmt.Errorf("failed to create database: %w", err)
				}
			}

			if pg.Postgres == nil {
				pg.Postgres = &provisioner.PostgresResource{}
			}
			dsn := dsn.DSN(dbName, dsn.Port(postgresPort))
			pg.Postgres.Output = &provisioner.PostgresResource_PostgresResourceOutput{
				WriteDsn: dsn,
				ReadDsn:  dsn,
			}
			return rc.Resource, nil
		},
	})
}
