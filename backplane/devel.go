package backplane

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/TBD54566975/ftl/backplane/internal/sql"
)

func CreateForDevel(ctx context.Context, dsn string, recreate bool) (*pgx.Conn, error) {
	return sql.CreateForDevel(ctx, dsn, recreate)
}
