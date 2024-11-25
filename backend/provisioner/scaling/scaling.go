package scaling

import (
	"context"
	"net/url"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/schema"
)

type RunnerScaling interface {
	Start(ctx context.Context) error

	GetEndpointForDeployment(ctx context.Context, module string, deployment string) (optional.Option[url.URL], error)

	StartDeployment(module string, deployment string, language *schema.Module) error

	TerminateDeployment(module string, deployment string) error
}
