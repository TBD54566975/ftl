package dao

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backplane/internal/sql/sqltest"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/schema"
)

func TestDAO(t *testing.T) {
	conn := sqltest.OpenForTesting(t)
	bp := New(conn)
	assert.NotZero(t, bp)

	ctx := context.Background()

	err := bp.CreateModule(ctx, "go", "test")
	assert.NoError(t, err)

	testSha, err := bp.CreateArtefact(ctx, []byte("test"))
	assert.NoError(t, err)

	module := &schema.Module{Name: "test"}
	key, err := bp.CreateDeployment(ctx, "go", module, []DeploymentArtefact{{
		Digest:     testSha,
		Executable: true,
		Path:       "dir/filename",
	}})
	assert.NoError(t, err)

	testSHA := sha256.MustParseSHA256("9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08")
	expected := &Deployment{
		Module:   "test",
		Language: "go",
		Schema:   module,
		Key:      key,
		Artefacts: []*Artefact{
			{Path: "dir/filename",
				Executable: true,
				Digest:     testSHA,
				Content:    []byte("test")},
		},
	}

	actual, err := bp.GetLatestDeployment(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	actual, err = bp.GetDeployment(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	misshingSHA := sha256.MustParseSHA256("fae7e4cbdca7167bbea4098c05d596f50bbb18062b61c1dfca3705b4a6c2888c")
	missing, err := bp.GetMissingArtefacts(ctx, []sha256.SHA256{testSHA, misshingSHA})
	assert.NoError(t, err)
	assert.Equal(t, []sha256.SHA256{misshingSHA}, missing)
}
