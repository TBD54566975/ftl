package rpc

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

// DefaultClient is the default HTTP client used internally by FTL.
var DefaultClient = func() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return &http.Client{
		// Timeout:   time.Second * 10,
		Transport: netTransport,
	}
}()

type PingableClient interface {
	Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

type ClientFactory[Client PingableClient] func(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) Client

func Dial[Client PingableClient](factory ClientFactory[Client], baseURL string, opts ...connect.ClientOption) Client {
	opts = append(opts, DefaultClientOptions()...)
	return factory(DefaultClient, baseURL, opts...)
}

// Middleware to inject values from a Context into the request Context.
func Middleware(ctx context.Context, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(mergedContext{values: ctx, Context: r.Context()})
		handler.ServeHTTP(w, r)
	}
}

var _ context.Context = (*mergedContext)(nil)

type mergedContext struct {
	values context.Context
	context.Context
}

func (m mergedContext) Value(key any) any {
	if value := m.Context.Value(key); value != nil {
		return value
	}
	return m.values.Value(key)
}
