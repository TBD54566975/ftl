package ingress

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/cors"
	ftlhttp "github.com/TBD54566975/ftl/internal/http"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/routing"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
	"github.com/TBD54566975/ftl/internal/slices"
)

type Config struct {
	Bind         *url.URL   `help:"Socket to bind to for ingress." default:"http://127.0.0.1:8891" env:"FTL_BIND"`
	AllowOrigins []*url.URL `help:"Allow CORS requests to ingress endpoints from these origins." env:"FTL_INGRESS_ALLOW_ORIGIN"`
	AllowHeaders []string   `help:"Allow these headers in CORS requests. (Requires AllowOrigins)" env:"FTL_INGRESS_ALLOW_HEADERS"`
}

func (c *Config) Validate() error {
	if len(c.AllowHeaders) > 0 && len(c.AllowOrigins) == 0 {
		return fmt.Errorf("AllowOrigins must be set when AllowHeaders is used")
	}
	return nil
}

type service struct {
	// Complete schema synchronised from the database.
	view   *atomic.Value[materialisedView]
	client routing.CallClient
}

// Start the HTTP ingress service. Blocks until the context is cancelled.
func Start(ctx context.Context, config Config, schemaEventSource schemaeventsource.EventSource, client routing.CallClient) error {
	logger := log.FromContext(ctx).Scope("http-ingress")
	ctx = log.ContextWithLogger(ctx, logger)
	svc := &service{
		view:   syncView(ctx, schemaEventSource),
		client: client,
	}

	ingressHandler := otelhttp.NewHandler(http.Handler(svc), "ftl.ingress")
	if len(config.AllowOrigins) > 0 {
		ingressHandler = cors.Middleware(
			slices.Map(config.AllowOrigins, func(u *url.URL) string { return u.String() }),
			config.AllowHeaders,
			ingressHandler,
		)
	}

	// Start the HTTP server
	logger.Infof("HTTP ingress server listening on: %s", config.Bind)
	err := ftlhttp.Serve(ctx, config.Bind, ingressHandler)
	if err != nil {
		return fmt.Errorf("ingress service stopped: %w", err)
	}
	return nil
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		return
	}
	start := time.Now()
	method := strings.ToLower(r.Method)
	requestKey := model.NewRequestKey(model.OriginIngress, fmt.Sprintf("%s %s", method, r.URL.Path))

	state := s.view.Load()
	routes := state.routes[r.Method]
	if len(routes) == 0 {
		http.NotFound(w, r)
		metrics.Request(r.Context(), r.Method, r.URL.Path, optional.None[*schemapb.Ref](), start, optional.Some("route not found in dal"))
		return
	}
	handleHTTP(start, state.schema, requestKey, routes, w, r, s.client)
}
