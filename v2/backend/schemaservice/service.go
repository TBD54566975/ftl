package schemaservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/pubsub"

	"github.com/TBD54566975/ftl"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	ftlpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v2alpha1"
	ftlpbconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v2alpha1/v2alpha1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

type Config struct {
	Bind                *url.URL             `help:"The address to bind the service to." default:"http://127.0.0.1:9992"`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11n-"`
}

//sumtype:decl
type deploymentChange interface{ change() }

type deploymentChanged struct {
	module *schema.Module
}

func (deploymentChanged) change() {}

type deploymentDeleted struct {
	module *schema.Module
}

func (deploymentDeleted) change() {}

type Service struct {
	state   *State
	changes *pubsub.Topic[deploymentChange]
}

func Start(ctx context.Context, config Config) error {
	logger := log.FromContext(ctx).Scope("schemaservice")
	ctx, doneFunc := context.WithCancel(ctx)
	defer doneFunc()
	err := observability.Init(ctx, false, "", "ftl-schema", ftl.Version, config.ObservabilityConfig)
	if err != nil {
		return fmt.Errorf("failed to initialise observability: %w", err)
	}
	svc := &Service{
		changes: pubsub.New[deploymentChange](),
		state:   NewState(),
	}
	logger.Debugf("Starting SchemaService on %s", config.Bind)
	err = rpc.Serve(ctx, config.Bind,
		rpc.GRPC(ftlpbconnect.NewSchemaServiceHandler, svc),
		rpc.HealthCheck(svc.healthCheck),
	)
	if err != nil {
		return fmt.Errorf("failed to start SchemaService: %w", err)
	}
	return nil
}

var _ ftlpbconnect.SchemaServiceHandler = (*Service)(nil)

func (s *Service) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) GetSchema(ctx context.Context, req *connect.Request[ftlpb.GetSchemaRequest]) (*connect.Response[ftlpb.GetSchemaResponse], error) {
	pbschema := s.state.Schema().ToProto().(*schemapb.Schema) //nolint:forcetypeassert
	return connect.NewResponse(&ftlpb.GetSchemaResponse{Schema: pbschema}), nil
}

// DeleteModule implements v2alpha1connect.SchemaServiceHandler.
func (s *Service) DeleteModule(ctx context.Context, req *connect.Request[ftlpb.DeleteModuleRequest]) (*connect.Response[ftlpb.DeleteModuleResponse], error) {
	module, err := s.state.DeleteModule(req.Msg.ModuleName)
	if errors.Is(err, ErrNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("module %q not found", req.Msg.ModuleName))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	s.changes.Publish(deploymentDeleted{module: module})
	return connect.NewResponse(&ftlpb.DeleteModuleResponse{}), nil
}

func (s *Service) UpsertModule(ctx context.Context, req *connect.Request[ftlpb.UpsertModuleRequest]) (*connect.Response[ftlpb.UpsertModuleResponse], error) {
	moduleSchema, err := schema.ModuleFromProto(req.Msg.Schema)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := s.state.UpsertModule(moduleSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	// TODO: this could be a synchronous publish/subscribe. I'm not 100% sure we want this though.
	s.changes.Publish(deploymentChanged{module: moduleSchema})
	return connect.NewResponse(&ftlpb.UpsertModuleResponse{}), nil
}

func (s *Service) PullSchema(ctx context.Context, req *connect.Request[ftlpb.PullSchemaRequest], resp *connect.ServerStream[ftlpb.PullSchemaResponse]) error {
	sub := s.changes.Subscribe(make(chan deploymentChange, 128))
	defer s.changes.Unsubscribe(sub)

	// Send initial state.
	responses := s.stateToResponses()
	for _, response := range responses {
		if err := resp.Send(response); err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
	}

	// Start listening for and sending updates.
	for {
		var err error
		select {
		case <-ctx.Done():
			return connect.NewError(connect.CodeCanceled, fmt.Errorf("pull schema cancelled: %w", ctx.Err()))

		case change := <-sub:
			switch change := change.(type) {
			case deploymentChanged:
				err = resp.Send(&ftlpb.PullSchemaResponse{ //nolint:forcetypeassert
					Schema: change.module.ToProto().(*schemapb.Module),
				})

			case deploymentDeleted:
				err = resp.Send(&ftlpb.PullSchemaResponse{ //nolint:forcetypeassert
					Deleted: true,
					Schema:  change.module.ToProto().(*schemapb.Module),
				})
			}
		}
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
	}
}

func (s *Service) stateToResponses() []*ftlpb.PullSchemaResponse {
	sch := s.state.Schema()
	var responses []*ftlpb.PullSchemaResponse
	for i, module := range sch.Modules {
		responses = append(responses, &ftlpb.PullSchemaResponse{ //nolint:forcetypeassert
			Schema:       module.ToProto().(*schemapb.Module),
			InitialBatch: i < len(sch.Modules)-1,
		})
	}
	return responses
}
