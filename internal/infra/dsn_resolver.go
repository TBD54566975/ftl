package infra

import (
	"context"
	"fmt"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

func ResolvePostgresDSN(ctx context.Context, schema *schemapb.DatabaseRuntime) (string, error) {
	dsnRuntime, ok := schema.Value.(*schemapb.DatabaseRuntime_DsnDatabaseRuntime)
	if !ok {
		return "", fmt.Errorf("unexpected database runtime type: %T", schema.Value)
	}
	return dsnRuntime.DsnDatabaseRuntime.Dsn, nil
}

func ResolveMySQLDSN(ctx context.Context, schema *schemapb.DatabaseRuntime) (string, error) {
	dsnRuntime, ok := schema.Value.(*schemapb.DatabaseRuntime_DsnDatabaseRuntime)
	if !ok {
		return "", fmt.Errorf("unexpected database runtime type: %T", schema.Value)
	}
	return dsnRuntime.DsnDatabaseRuntime.Dsn, nil
}
