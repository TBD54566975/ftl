package rpc

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/common/log"
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

type Pingable interface {
	Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

type ClientFactory[Client Pingable] func(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) Client

func Dial[Client Pingable](factory ClientFactory[Client], baseURL string, opts ...connect.ClientOption) Client {
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

// Wait for a client to become available.
//
// This will repeatedly call Ping() every 100ms until the service becomes
// available. TODO: This will probably need to be smarter at some point.
//
// If "ctx" is cancelled this will return ctx.Err()
func Wait(ctx context.Context, client Pingable) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		default:
		}
		_, err := client.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
		if err != nil {
			logger.Tracef("Ping failed waiting for client: %+v", err)
			time.Sleep(time.Millisecond * 100)
			continue
		}
		return err
	}
}
