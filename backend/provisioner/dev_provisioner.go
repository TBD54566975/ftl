package provisioner

import (
	"context"
	"fmt"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"

	"github.com/TBD54566975/ftl/backend/controller/dsn"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/internal/dev"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

// NewDevProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode
func NewDevProvisioner(postgresPort int, mysqlPort int) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[ResourceType]InMemResourceProvisionerFn{
		ResourceTypePostgres: provisionPostgres(postgresPort),
		ResourceTypeMysql:    provisionMysql(mysqlPort),
	})
}

func provisionMysql(mysqlPort int) InMemResourceProvisionerFn {
	return func(ctx context.Context, rc *provisioner.ResourceContext, module, id string) (*provisioner.Resource, error) {
		mysql, ok := rc.Resource.Resource.(*provisioner.Resource_Mysql)
		if !ok {
			panic(fmt.Errorf("unexpected resource type: %T", rc.Resource.Resource))
		}
		logger := log.FromContext(ctx)
		logger.Infof("provisioning mysql database: %s_%s", module, id)

		dbName := strcase.ToLowerCamel(module) + "_" + strcase.ToLowerCamel(id)

		// We assume that the DB hsas already been started when running in dev mode
		mysqlDSN, err := dev.SetupMySQL(ctx, "mysql:8.4.3", mysqlPort)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for mysql to be ready: %w", err)
		}
		conn, err := otelsql.Open("mysql", mysqlDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to mysql: %w", err)
		}
		defer conn.Close()

		res, err := conn.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to query database: %w", err)
		}
		defer res.Close()
		if res.Next() {
			_, err = conn.ExecContext(ctx, "DROP DATABASE "+dbName)
			if err != nil {
				return nil, fmt.Errorf("failed to drop database: %w", err)
			}
		}

		_, err = conn.ExecContext(ctx, "CREATE DATABASE "+dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}

		if mysql.Mysql == nil {
			mysql.Mysql = &provisioner.MysqlResource{}
		}
		dsn := dsn.MySQLDSN(dbName, dsn.Port(mysqlPort))
		mysql.Mysql.Output = &provisioner.MysqlResource_MysqlResourceOutput{
			WriteDsn: dsn,
			ReadDsn:  dsn,
		}
		return rc.Resource, nil
	}
}

func provisionPostgres(postgresPort int) func(ctx context.Context, rc *provisioner.ResourceContext, module string, id string) (*provisioner.Resource, error) {
	return func(ctx context.Context, rc *provisioner.ResourceContext, module, id string) (*provisioner.Resource, error) {
		pg, ok := rc.Resource.Resource.(*provisioner.Resource_Postgres)
		if !ok {
			panic(fmt.Errorf("unexpected resource type: %T", rc.Resource.Resource))
		}
		logger := log.FromContext(ctx)
		logger.Infof("provisioning postgres database: %s_%s", module, id)

		dbName := strcase.ToLowerCamel(module) + "_" + strcase.ToLowerCamel(id)

		// We assume that the DB has already been started when running in dev mode
		postgresDSN, err := dev.WaitForPostgresReady(ctx, postgresPort)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for postgres to be ready: %w", err)
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
		if res.Next() {
			// Terminate any dangling connections.
			_, err = conn.ExecContext(ctx, `
			SELECT pid, pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = $1 AND pid <> pg_backend_pid()`,
				dbName)
			if err != nil {
				return nil, fmt.Errorf("failed to kill existing backends: %w", err)
			}
			_, err = conn.ExecContext(ctx, "DROP DATABASE "+dbName)
			if err != nil {
				return nil, fmt.Errorf("failed to create database: %w", err)
			}
		}
		_, err = conn.ExecContext(ctx, "CREATE DATABASE "+dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}

		if pg.Postgres == nil {
			pg.Postgres = &provisioner.PostgresResource{}
		}
		dsn := dsn.PostgresDSN(dbName, dsn.Port(postgresPort))
		pg.Postgres.Output = &provisioner.PostgresResource_PostgresResourceOutput{
			WriteDsn: dsn,
			ReadDsn:  dsn,
		}
		return rc.Resource, nil
	}
}
