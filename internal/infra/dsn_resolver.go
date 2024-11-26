package infra

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/internal/schema"
)

func ResolvePostgresDSN(ctx context.Context, connector schema.DatabaseConnector) (string, error) {
	dsnRuntime, ok := connector.(*schema.DSNDatabaseConnector)
	if !ok {
		return "", fmt.Errorf("unexpected database connector type: %T", connector)
	}
	return dsnRuntime.DSN, nil
}

func ResolveMySQLDSN(ctx context.Context, connector schema.DatabaseConnector) (string, error) {
	dsnRuntime, ok := connector.(*schema.DSNDatabaseConnector)
	if !ok {
		return "", fmt.Errorf("unexpected database connector type: %T", connector)
	}
	return dsnRuntime.DSN, nil
}
