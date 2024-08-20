package observability

import (
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func OpenDBAndInstrument(dsn string) (*sql.DB, error) {
	conn, err := otelsql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = otelsql.RegisterDBStatsMetrics(conn, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return nil, fmt.Errorf("failed to register database metrics: %w", err)
	}

	return conn, nil
}
