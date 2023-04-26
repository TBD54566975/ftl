package dao

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/backplane/internal/sql/sqltest"
	"github.com/TBD54566975/ftl/schema"
)

func TestDAO(t *testing.T) {
	conn := sqltest.OpenForTesting(t)
	bp := New(conn)
	assert.NotZero(t, bp)

	ctx := context.Background()

	err := bp.CreateModule(ctx, "go", "test")
	assert.NoError(t, err)

	aid, err := bp.CreateArtefact(ctx, "dir/filename", true, []byte("test"))
	assert.NoError(t, err)

	module := &schema.Module{}
	err = bp.CreateDeployment(ctx, "test", module, []int64{aid})
	assert.NoError(t, err)

	actual, err := bp.GetLatestDeployment(ctx, "test")
	assert.NoError(t, err)

	actual.Key = uuid.UUID{}
	expected := &Deployment{
		Module:   "test",
		Language: "go",
		Schema:   module,
		Artefacts: []*Artefact{
			{Path: "dir/filename",
				Executable: true,
				Content:    []byte("test")},
		},
	}
	assert.Equal(t, expected, actual)
}
