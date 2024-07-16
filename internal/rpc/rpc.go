package rpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jpillora/backoff"
	"golang.org/x/net/http2"

	"github.com/TBD54566975/ftl/authn"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
)

// InitialiseClients initialises global HTTP clients used by the RPC system.
//
// "authenticators" are authenticator executables to use for each endpoint. The key is the URL of the endpoint, the
// value is the path to the authenticator executable.
//
// "allowInsecure" skips certificate verification, making TLS susceptible to machine-in-the-middle attacks.
func InitialiseClients(authenticators map[string]string, allowInsecure bool) {
	// We can't have a client-wide timeout because it also applies to
	// streaming RPCs, timing them out.
	h2cClient = &http.Client{
		Transport: authn.Transport(&http2.Transport{
			AllowHTTP: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: allowInsecure, // #nosec G402
			},
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				conn, err := dialer.Dial(network, addr)
				return conn, err
			},
		}, authenticators),
	}
	tlsClient = &http.Client{
		Transport: authn.Transport(&http2.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: allowInsecure, // #nosec G402
			},
			DialTLSContext: func(ctx context.Context, network, addr string, config *tls.Config) (net.Conn, error) {
				tlsDialer := tls.Dialer{Config: config, NetDialer: dialer}
				conn, err := tlsDialer.DialContext(ctx, network, addr)
				return conn, err
			},
		}, authenticators),
	}

	// Use a separate client for HTTP/1.1 with TLS.
	http1TLSClient = &http.Client{
		Transport: authn.Transport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: allowInsecure, // #nosec G402
			},
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				logger := log.FromContext(ctx)
				logger.Debugf("HTTP/1.1 connecting to %s %s", network, addr)

				tlsDialer := tls.Dialer{NetDialer: dialer}
				conn, err := tlsDialer.DialContext(ctx, network, addr)
				return conn, fmt.Errorf("HTTP/1.1 TLS dial failed: %w", err)
			},
		}, authenticators),
	}
}

func init() {
	InitialiseClients(map[string]string{}, false)
}

var (
	dialer = &net.Dialer{
		Timeout: time.Second * 10,
	}
	h2cClient *http.Client
	tlsClient *http.Client
	// Temporary client for HTTP/1.1 with TLS to help with debugging.
	http1TLSClient *http.Client
)

type Pingable interface {
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

// GetHTTPClient returns a HTTP client usable for the given URL.
func GetHTTPClient(url string) *http.Client {
	if h2cClient == nil {
		panic("rpc.InitialiseClients() must be called before GetHTTPClient()")
	}

	// TEMP_GRPC_HTTP1_ONLY set to non blank will use http1TLSClient
	if os.Getenv("TEMP_GRPC_HTTP1_ONLY") != "" {
		return http1TLSClient
	}

	if strings.HasPrefix(url, "http://") {
		return h2cClient
	}
	return tlsClient
}

// ClientFactory is a function that creates a new client and is typically one of
// the New*Client functions generated by protoc-gen-connect-go.
type ClientFactory[Client Pingable] func(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) Client

func Dial[Client Pingable](factory ClientFactory[Client], baseURL string, errorLevel log.Level, opts ...connect.ClientOption) Client {
	client := GetHTTPClient(baseURL)
	opts = append(opts, DefaultClientOptions(errorLevel)...)
	return factory(client, baseURL, opts...)
}

// ContextValuesMiddleware injects values from a Context into the request Context.
func ContextValuesMiddleware(ctx context.Context, handler http.Handler) http.HandlerFunc {
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
// ready. TODO: This will probably need to be smarter at some point.
//
// If "ctx" is cancelled this will return ctx.Err()
func Wait(ctx context.Context, retry backoff.Backoff, client Pingable) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		resp, err := client.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
		if err == nil {
			if resp.Msg.NotReady == nil {
				return nil
			}
			err = fmt.Errorf("service is not ready: %s", *resp.Msg.NotReady)
		}
		delay := retry.Duration()
		logger.Tracef("Ping failed waiting %s for client: %+v", delay, err)
		time.Sleep(delay)
	}
}

// RetryStreamingClientStream will repeatedly call handler with the stream
// returned by "rpc" until handler returns an error or the context is cancelled.
//
// If the stream errors, it will be closed and a new call will be issued.
func RetryStreamingClientStream[Req, Resp any](
	ctx context.Context,
	retry backoff.Backoff,
	rpc func(context.Context) *connect.ClientStreamForClient[Req, Resp],
	handler func(ctx context.Context, send func(*Req) error) error,
) {
	logLevel := log.Debug
	errored := false
	logger := log.FromContext(ctx)
	for {
		stream := rpc(ctx)
		var err error
		for {
			err = handler(ctx, stream.Send)
			if err != nil {
				break
			}
			if errored {
				logger.Debugf("Client stream recovered")
				errored = false
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
			retry.Reset()
			logLevel = log.Warn
		}

		// We've hit an error.
		_, closeErr := stream.CloseAndReceive()
		if closeErr != nil {
			logger.Logf(log.Debug, "Failed to close stream: %s", closeErr)
		}

		errored = true
		delay := retry.Duration()
		if !errors.Is(err, context.Canceled) {
			logger.Logf(logLevel, "Stream handler failed, retrying in %s: %s", delay, err)
		}
		select {
		case <-ctx.Done():
			return

		case <-time.After(delay):
		}

	}
}

// AlwaysRetry instructs RetryStreamingServerStream to always retry the errors it encounters when
// supplied as the errorRetryCallback argument
func AlwaysRetry() func(error) bool {
	return func(err error) bool { return true }
}

// RetryStreamingServerStream will repeatedly call handler with responses from
// the stream returned by "rpc" until handler returns an error or the context is
// cancelled.
func RetryStreamingServerStream[Req, Resp any](
	ctx context.Context,
	retry backoff.Backoff,
	req *Req,
	rpc func(context.Context, *connect.Request[Req]) (*connect.ServerStreamForClient[Resp], error),
	handler func(ctx context.Context, resp *Resp) error,
	errorRetryCallback func(err error) bool,
) {
	logLevel := log.Debug
	errored := false
	logger := log.FromContext(ctx)
	for {
		stream, err := rpc(ctx, connect.NewRequest(req))
		if err == nil {
			for {
				if stream.Receive() {
					resp := stream.Msg()
					err = handler(ctx, resp)

					if err != nil {
						break
					}
					if errored {
						logger.Debugf("Server stream recovered")
						errored = false
					}
					select {
					case <-ctx.Done():
						return
					default:
					}
					retry.Reset()
					logLevel = log.Warn
				} else {
					// Stream terminated; check if this was caused by an error
					err = stream.Err()
					logLevel = logLevelForError(err)
					break
				}
			}
		}

		errored = true
		delay := retry.Duration()
		if err != nil && !errors.Is(err, context.Canceled) {
			if errorRetryCallback != nil && !errorRetryCallback(err) {
				logger.Errorf(err, "Stream handler encountered a non-retryable error")
				return
			}

			logger.Logf(logLevel, "Stream handler failed, retrying in %s: %s", delay, err)
		} else if err == nil {
			logger.Debugf("Stream finished, retrying in %s", delay)
		}

		select {
		case <-ctx.Done():
			return

		case <-time.After(delay):
		}

	}
}

// useDebugErrorLevel indicates whether the specified error should be reported as a debug
// level log.
func logLevelForError(err error) log.Level {
	if err != nil && strings.Contains(err.Error(), "connect: connection refused") {
		return log.Debug
	}
	return log.Warn
}
