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

	StartDeployment(ctx context.Context, module string, deployment string, sch *schema.Module) error

	TerminateDeployment(ctx context.Context, module string, deployment string) error
}