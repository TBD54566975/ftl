package admin

import (
	"context"

	"github.com/TBD54566975/ftl/common/configuration"
)

// localClient reads and writes to local projectconfig files without making any network
// calls. It allows us to interface with local ftl-project.toml files without needing to
// start a controller.
type localClient struct {
	*AdminService
}

func newLocalClient(ctx context.Context) *localClient {
	cm := configuration.ConfigFromContext(ctx)
	sm := configuration.SecretsFromContext(ctx)
	return &localClient{NewAdminService(cm, sm)}
}
