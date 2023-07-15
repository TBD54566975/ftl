package controller

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl/controller/internal/sql"
)

func CreateForDevel(ctx context.Context, dsn string, recreate bool) (*pgxpool.Pool, error) {
	return sql.CreateForDevel(ctx, dsn, recreate)
}
