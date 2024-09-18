package timeline

import (
	"context"
	"encoding/json"

	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/types/optional"
)

type IngresEvent struct {
	DeploymentKey model.DeploymentKey
	RequestKey    model.RequestKey
}

// The internal JSON payload of an ingress event.
type eventIngressJSON struct {
	DurationMS int64                   `json:"duration_ms"`
	Request    json.RawMessage         `json:"request"`
	Response   json.RawMessage         `json:"response"`
	Error      optional.Option[string] `json:"error,omitempty"`
	Stack      optional.Option[string] `json:"stack,omitempty"`
}

func (s *Service) InsertIngress(ctx context.Context, ingress *IngresEvent) error {
	return libdal.TranslatePGError(s.db.InsertTimelineIngressEvent(ctx, sql.InsertTimelineIngressEventParams{
		DeploymentKey: ingress.DeploymentKey,
		RequestKey:    optional.Some(ingress.RequestKey.String()),
	}))
}
