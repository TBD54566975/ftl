package provisioner

import (
	"context"
	"fmt"

	"github.com/XSAM/otelsql"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/internal/dev"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

// NewDevProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode
func NewDevProvisioner(postgresPort int) *InMemProvisioner {
	var postgresDSN string
	return NewEmbeddedProvisioner(map[ResourceType]InMemResourceProvisionerFn{
		ResourceTypePostgres: func(ctx context.Context, resource *provisioner.Resource, module, id string, step *InMemProvisioningStep) {
			pg, ok := resource.Resource.(*provisioner.Resource_Postgres)
			if !ok {
				panic(fmt.Errorf("unexpected resource type: %T", resource.Resource))
			}
			logger := log.FromContext(ctx)
			logger.Infof("provisioning postgres database: %s_%s", module, id)

			defer step.Done.Store(true)
			if postgresDSN == "" {
				// We assume that the DB hsas already been started when running in dev mode
				dsn, err := dev.WaitForDBReady(ctx, postgresPort)
				if err != nil {
					step.Err = err
					return
				}
				postgresDSN = dsn
			}
			dbName := strcase.ToLowerSnake(module) + "_" + strcase.ToLowerSnake(id)
			conn, err := otelsql.Open("pgx", postgresDSN)
			if err != nil {
				step.Err = err
				return
			}
			defer conn.Close()

			res, err := conn.Query("SELECT * FROM pg_catalog.pg_database WHERE datname=$1", dbName)
			if err != nil {
				step.Err = err
				return
			}
			defer res.Close()
			if !res.Next() {
				_, err = conn.ExecContext(ctx, "CREATE DATABASE "+dbName)
				if err != nil {
					step.Err = err
					return
				}
			}

			if pg.Postgres == nil {
				pg.Postgres = &provisioner.PostgresResource{}
			}
			pg.Postgres.Output = &provisioner.PostgresResource_PostgresResourceOutput{
				ReadEndpoint:  postgresDSN,
				WriteEndpoint: postgresDSN,
			}
		},
	})
}
