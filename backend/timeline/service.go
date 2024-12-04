package timeline

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

type Config struct {
	Bind *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8894" env:"FTL_BIND"`
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c); err != nil {
		panic(err)
	}
}

type service struct {
	// TODO: add timeline schema view? (whatever is needed per the current timeline tables)
}

func Start(ctx context.Context, config Config, schemaEventSource schemaeventsource.EventSource) error {
	config.SetDefaults()

	logger := log.FromContext(ctx).Scope("timeline")
	svc := &service{}

	logger.Debugf("Timeline service listening on: %s", config.Bind)
	err := rpc.Serve(ctx, config.Bind,
		rpc.GRPC(ftlv1connect.NewTimelineServiceHandler, svc),
	)
	if err != nil {
		return fmt.Errorf("timeline service stopped serving: %w", err)
	}
	return nil
}

func (s *service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	panic("not implemented")
}

func (s *service) GetTimeline(ctx context.Context, req *connect.Request[ftlv1.GetTimelineRequest]) (*connect.Response[ftlv1.GetTimelineResponse], error) {
	panic("not implemented")
}

func (s *service) DeleteOldEvents(ctx context.Context, req *connect.Request[ftlv1.DeleteOldEventsRequest]) (*connect.Response[ftlv1.DeleteOldEventsResponse], error) {
	panic("not implemented")
}
