package ftl

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestSecret(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	sr := configuration.ProjectConfigResolver[configuration.Secrets]{Config: []string{"testdata/ftl-project.toml"}}
	sm, err := configuration.NewSecretsManager(ctx, sr)
	assert.NoError(t, err)
	ctx = configuration.ContextWithSecrets(ctx, sm)
	type C struct {
		One string
		Two string
	}
	config := Secret[C]("secret")
	assert.Equal(t, C{"one", "two"}, config.Get(ctx))
}
