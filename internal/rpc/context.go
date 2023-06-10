package rpc

import (
	"context"
	"net/http"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/observability"
	"github.com/TBD54566975/ftl/schema"
)

const ftlDirectRoutingHeader = "FTL-Direct"
const ftlVerbHeader = "FTL-Verb"

type ftlDirectRoutingKey struct{}
type ftlVerbKey struct{}

// WithDirectRouting ensures any hops in Verb routing do not redirect.
//
// This is used so that eg. calls from Drives do not create recursive loops
// when calling back to the Agent.
func WithDirectRouting(ctx context.Context) context.Context {
	return context.WithValue(ctx, ftlDirectRoutingKey{}, "1")
}

// WithVerb sets the module.verb of the current request.
func WithVerb(ctx context.Context, ref *schema.VerbRef) context.Context {
	return context.WithValue(ctx, ftlVerbKey{}, ref)
}

// VerbFromContext returns the module.verb of the current request.
func VerbFromContext(ctx context.Context) (*schema.VerbRef, bool) {
	value := ctx.Value(ftlVerbKey{})
	verb, ok := value.(*schema.VerbRef)
	return verb, ok
}

// IsDirectRouted returns true if the incoming request should be directly
// routed and never redirected.
func IsDirectRouted(ctx context.Context) bool {
	return ctx.Value(ftlDirectRoutingKey{}) != nil
}

func DefaultClientOptions(level log.Level) []connect.ClientOption {
	return []connect.ClientOption{
		connect.WithGRPC(), // Use gRPC because some servers will not be using Connect.
		connect.WithInterceptors(MetadataInterceptor(level)),
	}
}

func DefaultHandlerOptions() []connect.HandlerOption {
	return []connect.HandlerOption{
		connect.WithInterceptors(MetadataInterceptor(log.Error)),
		connect.WithInterceptors(otelconnect.NewInterceptor()),
		connect.WithInterceptors(observability.NewInterceptor()),
	}
}

// MetadataInterceptor propagates FTL metadata through servers and clients.
func MetadataInterceptor(level log.Level) connect.Interceptor {
	return &metadataInterceptor{
		errorLevel: level,
	}
}

type metadataInterceptor struct {
	errorLevel log.Level
}

func (*metadataInterceptor) WrapStreamingClient(req connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		// TODO(aat): I can't figure out how to get the client headers here.
		logger := log.FromContext(ctx)
		logger.Tracef("%s (streaming client)", s.Procedure)
		return req(ctx, s)
	}
}

func (m *metadataInterceptor) WrapStreamingHandler(req connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, s connect.StreamingHandlerConn) error {
		logger := log.FromContext(ctx)
		logger.Tracef("%s (streaming handler)", s.Spec().Procedure)
		ctx, err := updateContext(ctx, s.Spec().IsClient, s.RequestHeader())
		if err != nil {
			return err
		}
		err = errors.WithStack(req(ctx, s))
		if err != nil {
			logger.Logf(m.errorLevel, "Streaming RPC failed: %s: %s", err, s.Spec().Procedure)
			return err
		}
		return nil
	}
}

func (m *metadataInterceptor) WrapUnary(uf connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		logger := log.FromContext(ctx)
		logger.Tracef("%s (unary)", req.Spec().Procedure)
		ctx, err := updateContext(ctx, req.Spec().IsClient, req.Header())
		if err != nil {
			return nil, err
		}
		resp, err := uf(ctx, req)
		if err != nil {
			err = errors.WithStack(err)
			logger.Logf(m.errorLevel, "Unary RPC failed: %s: %s", err, req.Spec().Procedure)
			return nil, err
		}
		return resp, nil
	})
}

type clientKey[Client Pingable] struct{}

// ContextWithClient returns a context with an RPC client attached.
func ContextWithClient[Client Pingable](ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, clientKey[Client]{}, client)
}

// ClientFromContext returns the given RPC client from the context, or panics.
func ClientFromContext[Client Pingable](ctx context.Context) Client {
	value := ctx.Value(clientKey[Client]{})
	if value == nil {
		panic("no RPC client in context")
	}
	return value.(Client) //nolint:forcetypeassert
}

func updateContext(ctx context.Context, isClient bool, headers http.Header) (context.Context, error) {
	if isClient {
		if IsDirectRouted(ctx) {
			headers.Set(ftlDirectRoutingHeader, "1")
		}
		if verb, ok := VerbFromContext(ctx); ok {
			headers.Set(ftlVerbHeader, verb.String())
		}
	} else {
		if headers.Get(ftlDirectRoutingHeader) != "" {
			ctx = WithDirectRouting(ctx)
		}
		if verb := headers.Get(ftlVerbHeader); verb != "" {
			ref, err := schema.ParseRef(verb)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid verb reference %q", verb)
			}
			ctx = WithVerb(ctx, (*schema.VerbRef)(ref))
		}
	}
	return ctx, nil
}
