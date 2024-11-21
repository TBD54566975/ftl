package validation

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type Empty = ftl.Config[string]

//ftl:cron * * * * * 9999
func BadYear(ctx context.Context) error {
	return nil
}

//ftl:cron 0 0 0 0 0
func AllZeroes(ctx context.Context) error {
	return nil
}
