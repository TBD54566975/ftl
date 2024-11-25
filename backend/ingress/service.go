package ingress

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/controller/observability"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/cors"
	ftlhttp "github.com/TBD54566975/ftl/internal/http"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

type PullSchemaClient interface {
	PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest]) (*connect.ServerStreamForClient[ftlv1.PullSchemaResponse], error)
}

type CallClient interface {
	Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error)
}

type Config struct {
	Bind         *url.URL   `help:"Socket to bind to for ingress." default:"http://127.0.0.1:8891" env:"FTL_INGRESS_BIND"`
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
	schemaState atomic.Value[schemaState]
	callClient  CallClient
}

// Start the HTTP ingress service. Blocks until the context is cancelled.
func Start(ctx context.Context, config Config, pullSchemaClient PullSchemaClient, verbClient CallClient) error {
	wg, ctx := errgroup.WithContext(ctx)
	logger := log.FromContext(ctx).Scope("http-ingress")
	svc := &service{
		callClient: verbClient,
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
	wg.Go(func() error {
		logger.Infof("HTTP ingress server listening on: %s", config.Bind)
		return ftlhttp.Serve(ctx, config.Bind, ingressHandler)
	})
	// Start watching for schema changes.
	wg.Go(func() error {
		rpc.RetryStreamingServerStream(ctx, "pull-schema", backoff.Backoff{}, &ftlv1.PullSchemaRequest{}, pullSchemaClient.PullSchema, func(ctx context.Context, resp *ftlv1.PullSchemaResponse) error {
			existing := svc.schemaState.Load().protoSchema
			newState := schemaState{
				protoSchema: &schemapb.Schema{},
				httpRoutes:  make(map[string][]ingressRoute),
			}
			if resp.ChangeType != ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED {
				found := false
				if existing != nil {
					for i := range existing.Modules {
						if existing.Modules[i].Name == resp.ModuleName {
							newState.protoSchema.Modules = append(newState.protoSchema.Modules, resp.Schema)
							found = true
						} else {
							newState.protoSchema.Modules = append(newState.protoSchema.Modules, existing.Modules[i])
						}
					}
				}
				if !found {
					newState.protoSchema.Modules = append(newState.protoSchema.Modules, resp.Schema)
				}
			} else if existing != nil {
				for i := range existing.Modules {
					if existing.Modules[i].Name != resp.ModuleName {
						newState.protoSchema.Modules = append(newState.protoSchema.Modules, existing.Modules[i])
					}
				}
			}

			newState.httpRoutes = extractIngressRoutingEntries(newState.protoSchema)
			sch, err := schema.FromProto(newState.protoSchema)
			if err != nil {
				// Not much we can do here, we don't update the state with the broken schema.
				logger.Errorf(err, "failed to parse schema")
				return nil
			}
			newState.schema = sch
			svc.schemaState.Store(newState)
			return nil
		}, rpc.AlwaysRetry())
		return nil
	})
	err := wg.Wait()
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

	routes := s.schemaState.Load().httpRoutes[r.Method]
	if len(routes) == 0 {
		http.NotFound(w, r)
		observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.None[*schemapb.Ref](), start, optional.Some("route not found in dal"))
		return
	}
	handleHTTP(start, s.schemaState.Load().schema, requestKey, routes, w, r, s.callClient)
}

type schemaState struct {
	protoSchema *schemapb.Schema
	schema      *schema.Schema
	httpRoutes  map[string][]ingressRoute
}

type ingressRoute struct {
	path   string
	module string
	verb   string
	method string
}

func extractIngressRoutingEntries(schema *schemapb.Schema) map[string][]ingressRoute {
	var ingressRoutes = make(map[string][]ingressRoute)
	for _, module := range schema.Modules {
		for _, decl := range module.Decls {
			if verb, ok := decl.Value.(*schemapb.Decl_Verb); ok {
				for _, metadata := range verb.Verb.Metadata {
					if ingress, ok := metadata.Value.(*schemapb.Metadata_Ingress); ok {
						ingressRoutes[ingress.Ingress.Method] = append(ingressRoutes[ingress.Ingress.Method], ingressRoute{
							verb:   verb.Verb.Name,
							method: ingress.Ingress.Method,
							path:   ingressPathString(ingress.Ingress.Path),
							module: module.Name,
						})
					}
				}
			}
		}
	}
	return ingressRoutes
}

func ingressPathString(path []*schemapb.IngressPathComponent) string {
	pathString := make([]string, len(path))
	for i, p := range path {
		switch p.Value.(type) {
		case *schemapb.IngressPathComponent_IngressPathLiteral:
			pathString[i] = p.GetIngressPathLiteral().Text
		case *schemapb.IngressPathComponent_IngressPathParameter:
			pathString[i] = fmt.Sprintf("{%s}", p.GetIngressPathParameter().Name)
		}
	}
	return "/" + strings.Join(pathString, "/")
}
