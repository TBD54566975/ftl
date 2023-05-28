package dal

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/controlplane/internal/sql/sqltest"
	"github.com/TBD54566975/ftl/schema"
)

func TestDAL(t *testing.T) {
	var testContent = bytes.Repeat([]byte("sometestcontentthatislongerthanthereadbuffer"), 100)
	var testSHA = sha256.Sum(testContent)

	conn := sqltest.OpenForTesting(t)
	bp := New(conn)
	assert.NotZero(t, bp)

	ctx := context.Background()

	err := bp.CreateModule(ctx, "go", "test")
	assert.NoError(t, err)

	testSha, err := bp.CreateArtefact(ctx, testContent)
	assert.NoError(t, err)

	module := &schema.Module{Name: "test"}
	key, err := bp.CreateDeployment(ctx, "go", module, []DeploymentArtefact{{
		Digest:     testSha,
		Executable: true,
		Path:       "dir/filename",
	}})
	assert.NoError(t, err)

	expected := &Deployment{
		Module:   "test",
		Language: "go",
		Schema:   module,
		Key:      key,
		Artefacts: []*Artefact{
			{Path: "dir/filename",
				Executable: true,
				Digest:     testSHA,
				Content:    bytes.NewReader(testContent)},
		},
	}
	expectedContent := artefactContent(t, expected.Artefacts)

	actual, err := bp.GetLatestDeployment(ctx, "test")
	assert.NoError(t, err)
	actualContent := artefactContent(t, actual.Artefacts)
	assert.Equal(t, expectedContent, actualContent)
	assert.Equal(t, expected, actual)

	actual, err = bp.GetDeployment(ctx, key)
	assert.NoError(t, err)
	actualContent = artefactContent(t, actual.Artefacts)
	assert.Equal(t, expectedContent, actualContent)
	assert.Equal(t, expected, actual)

	misshingSHA := sha256.MustParseSHA256("fae7e4cbdca7167bbea4098c05d596f50bbb18062b61c1dfca3705b4a6c2888c")
	missing, err := bp.GetMissingArtefacts(ctx, []sha256.SHA256{testSHA, misshingSHA})
	assert.NoError(t, err)
	assert.Equal(t, []sha256.SHA256{misshingSHA}, missing)
}

func artefactContent(t testing.TB, artefacts []*Artefact) [][]byte {
	t.Helper()
	var result [][]byte
	for _, a := range artefacts {
		content, err := io.ReadAll(a.Content)
		assert.NoError(t, err)
		result = append(result, content)
		a.Content = nil
	}
	return result
}
