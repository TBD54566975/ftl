package identity

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/cronjobs/dal"
	parentdal "github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestIdentity(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	conn := sqltest.OpenForTesting(ctx, t)
	dal := dal.New(conn)

	uri := "fake-kms://CK6YwYkBElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEJy4TIQgfCuwxA3ZZgChp_wYARABGK6YwYkBIAE"
	encryption, err := encryption.New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
	assert.NoError(t, err)

	parentDAL := parentdal.New(ctx, conn, encryption)

}
