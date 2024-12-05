package artefacts

import (
	"context"

	"connectrpc.com/connect"

	"github.com/TBD54566975/ftl/backend/internal/artefacts"
	artefactspb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/artefacts/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/artefacts/v1/artefactspbconnect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
)

type Service struct {
	registry *artefacts.OCIArtefactService
}

func New(registry *artefacts.OCIArtefactService) *Service {
	return &Service{registry: registry}
}

var _ artefactspbconnect.ArtefactsServiceHandler = (*Service)(nil)

func (s *Service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.Request[artefactspb.UploadArtefactRequest]) (*connect.Response[artefactspb.UploadArtefactResponse], error) {
	logger := log.FromContext(ctx)
	digest, err := s.registry.Upload(ctx, artefacts.Artefact{Content: req.Msg.Content})
	if err != nil {
		return nil, err
	}
	logger.Debugf("Created new artefact %s", digest)
	return connect.NewResponse(&artefactspb.UploadArtefactResponse{Digest: digest[:]}), nil
}
