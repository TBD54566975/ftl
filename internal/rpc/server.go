package rpc

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/alecthomas/concurrency"
	"github.com/alecthomas/types/pubsub"
)

const ShutdownGracePeriod = time.Second * 5

type serverOptions struct {
	mux             *http.ServeMux
	reflectionPaths []string
}

type Option func(*serverOptions)

type GRPCServerConstructor[Iface Pingable] func(svc Iface, opts ...connect.HandlerOption) (string, http.Handler)
type RawGRPCServerConstructor[Iface any] func(svc Iface, opts ...connect.HandlerOption) (string, http.Handler)

// GRPC is a convenience function for registering a GRPC server with default options.
// TODO(aat): Do we need pingable here?
func GRPC[Iface, Impl Pingable](constructor GRPCServerConstructor[Iface], impl Impl, options ...connect.HandlerOption) Option {
	return func(o *serverOptions) {
		options = append(options, DefaultHandlerOptions()...)
		path, handler := constructor(any(impl).(Iface), options...)
		o.reflectionPaths = append(o.reflectionPaths, strings.Trim(path, "/"))
		o.mux.Handle(path, handler)
	}
}

// RawGRPC is a convenience function for registering a GRPC server with default options without Pingable.
func RawGRPC[Iface, Impl any](constructor RawGRPCServerConstructor[Iface], impl Impl, options ...connect.HandlerOption) Option {
	return func(o *serverOptions) {
		options = append(options, DefaultHandlerOptions()...)
		path, handler := constructor(any(impl).(Iface), options...)
		o.reflectionPaths = append(o.reflectionPaths, strings.Trim(path, "/"))
		o.mux.Handle(path, handler)
	}
}

// HTTP adds a HTTP route to the server.
func HTTP(prefix string, handler http.Handler) Option {
	return func(o *serverOptions) {
		o.mux.Handle(prefix, handler)
	}
}

type Server struct {
	listen *url.URL
	Bind   *pubsub.Topic[*url.URL] // Will be updated with the actual bind address.
	Server *http.Server
}

func NewServer(ctx context.Context, listen *url.URL, options ...Option) (*Server, error) {
	opts := &serverOptions{
		mux: http.NewServeMux(),
	}

	opts.mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, option := range options {
		option(opts)
	}

	// Register reflection services.
	reflector := grpcreflect.NewStaticReflector(opts.reflectionPaths...)
	opts.mux.Handle(grpcreflect.NewHandlerV1(reflector))
	opts.mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	root := ContextValuesMiddleware(ctx, opts.mux)

	http1Server := &http.Server{
		Handler:           root,
		ReadHeaderTimeout: time.Second * 30,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	return &Server{
		listen: listen,
		Bind:   pubsub.New[*url.URL](),
		Server: http1Server,
	}, nil
}

// Serve runs the server, updating .Bind with the actual bind address.
func (s *Server) Serve(ctx context.Context) error {
	pemCertificate := `-----BEGIN CERTIFICATE----- 
MIIDhzCCAm+gAwIBAgIURimi3QEIsYsAR16pPrKi2+0C72EwDQYJKoZIhvcNAQEL
BQAwUzELMAkGA1UEBhMCVVMxDzANBgNVBAgMBkRlbmlhbDEMMAoGA1UEBwwDVEJE
MQwwCgYDVQQKDANUQkQxFzAVBgNVBAMMDmZ0bC50YmRkZXYub3JnMB4XDTI0MDYw
NjE4MjYwMVoXDTM0MDYwNDE4MjYwMVowUzELMAkGA1UEBhMCVVMxDzANBgNVBAgM
BkRlbmlhbDEMMAoGA1UEBwwDVEJEMQwwCgYDVQQKDANUQkQxFzAVBgNVBAMMDmZ0
bC50YmRkZXYub3JnMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0x8z
vSo4Iq9TNhvqmaSCIiK0JpYcdl7GnB3oyxa22ekuXMqNvhx2tQkDHN580LB+vVyg
8l9TnW5rzNU5PHKYzVtKv6+jSE08tv4wN/R5gZG7wc+zjnnBDmpiueecu/ZOpuAm
sPpzOjGR92ughBUilJ683hifIZvT0bTIx+NmsePAbAjhz+fHsjnEnX35roS6TWdK
37hJ6Wj01AzpBcrv5Gh7At9H0Pq0mq7bxsk6mA12TGOebwj4sANIpbsXOqp6X3Y/
b+rXJxsKbh2mqGB5qLlB2mtyFE6B7ldpSdtTQGFXm19TBWZ62ggGLoepU3atcpr8
xKIRH+HYo+bFb89DtwIDAQABo1MwUTAdBgNVHQ4EFgQUusYsmpf8oviHkHnEkVz+
UhqRz34wHwYDVR0jBBgwFoAUusYsmpf8oviHkHnEkVz+UhqRz34wDwYDVR0TAQH/
BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEABf4AqFF8CBcrZBcOvM8MVoiwXIaq
XFmWtwRrd5U3K89ZuaL9JgBvk851tYq6HMLehe6hJVypPqnCRkojX6ns75nP5wIk
mAOC+TTcJnvPTO8A7S2mSo8tv8NHX+956vrSLyGg/+9axxodCJDOIUp2flBMoMOU
wOEttjwPBIDBP+ix9XSMKkOFIshc3OfAulxNWR77FE4bhiGM/EzmKu2sajdssyjm
sH2BijCqbYNdDg1mhR9AUsPJwhSOnESEbY0itBCE7DKCnn5QHySUDIHIENhwk2pZ
40Go8SwDdHrrgKE4+8eMNUPr4Abbyzl5FiRXgIj+17CgiIE922J2Eq23KA==
-----END CERTIFICATE-----`

	pemKey := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDTHzO9Kjgir1M2
G+qZpIIiIrQmlhx2XsacHejLFrbZ6S5cyo2+HHa1CQMc3nzQsH69XKDyX1OdbmvM
1Tk8cpjNW0q/r6NITTy2/jA39HmBkbvBz7OOecEOamK555y79k6m4Caw+nM6MZH3
a6CEFSKUnrzeGJ8hm9PRtMjH42ax48BsCOHP58eyOcSdffmuhLpNZ0rfuEnpaPTU
DOkFyu/kaHsC30fQ+rSartvGyTqYDXZMY55vCPiwA0iluxc6qnpfdj9v6tcnGwpu
HaaoYHmouUHaa3IUToHuV2lJ21NAYVebX1MFZnraCAYuh6lTdq1ymvzEohEf4dij
5sVvz0O3AgMBAAECggEABXowXoBl9lOEBVET/B+v4btkbTpMGb3bnurtu0XKpPyL
kW+5bAW9zZDp5rRzNxeBlO3Fc2mmRk7WsCN5lg9HSCUU6Ytu724a494NJ957I2qX
ALdKnBq6N8Y/lgSVczubR43WakAGDkhk1j1m+fnVSTrOK+57fKTJtKooVHtFIPdQ
zBz607DwsAraBHPWGBVsK3wffZkDE58nCzoKEi7gDsbWqZg5tV2E4sKLQBbZpxYH
OxUUZVtcmZ3u3vkhrwPiMHV4heB29BFUY2GoChe+nRFKFZoJhfyweUzl3R+RRC9r
QPV5e32mzjjBuzUlfNV9RKYwdStJWvWMsMRRUgzUZQKBgQDq6uIEMMXvtntriLss
touyK4m5vgrGkoSuYZ5Dgze+nUMcX1D0y8CD5kvhWHeL2i5hB8k99U+WpjxLCBEa
OFw7c6kt7hkIBJKB60mbqsrtssfhA07fR3ZUZlDijHV0fkN918nRYQXsKKsJ677R
HNSAjmVXN8HU7YVZ8EneQtZshQKBgQDmEaBHeb5sSop8jTVypp/ElcFhurVmVfZR
YYxWh/Hefsczh3naKaeP4JzcLDTdCGo9ZK0xNUM2tmkGKXQfGRKMy2oEWO06mVgO
8wZiigz850LEvSm0Lgjh+ZySXQ5ROU9cn0fQ6SO1zMXtTUirNu4Ctng/I16LxkEU
A1oStV5SCwKBgCa5LyaHr6kTCIcyU8BMGvz0plBC3l3bOxnPp5nzYFYAcFaV869W
gtZ7ONjdj18zSN/fu7GF5Wes4VVw7/jFf5ahOysCC4hB0LCvy0NoxOinxsD1naO6
kOvarcyaYKYiRhfRYUgtWR+TmJYbESpBOVoznsrguwfRW2D29gY4OEZNAoGBAJE3
UumKiI0lx5+yKahCT9nvhG5BQTpky+K2JbSAfkQn1WhK/LidTixcY+X86SkSpKw3
nbHPoqsoG8ZN6AOw+apwwmwYDTTNkW1uK/uKk4QWHGi91VLrM6Qev5sKrXzLJbKa
vuO4JFgd9lhATbv0IesIbYG8u3KSIoWVUAc6/1vdAoGARegpq5MA3w+ruLmPXNGD
Ep1Hs9wyEIUDlFcId48SrXpXAbkeFZpBnq8DQ6Dm9ZX4UtX6L6vCvZkrkHZgF5nR
4SzkYcQX/XgR66wMbT+TkmkQqRI7S6ckACE23ibzcCkcjq7NZhzY2ecpOuZK6RCS
k4U2TWpjILEHjFHJIx+1+Gs=
-----END PRIVATE KEY-----`

	// Hardcoded certificate for testing
	certificate, err := tls.X509KeyPair([]byte(pemCertificate), []byte(pemKey))
	if err != nil {
		return err
	}
	certlist := []tls.Certificate{certificate}

	listener, err := tls.Listen("tcp", s.listen.Host, &tls.Config{Certificates: certlist})
	if err != nil {
		return err
	}
	if s.listen.Port() == "0" {
		s.listen.Host = listener.Addr().String()
	}
	s.Bind.Publish(s.listen)

	tree, _ := concurrency.New(ctx)

	// Shutdown server on context cancellation.
	tree.Go(func(ctx context.Context) error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), ShutdownGracePeriod)
		defer cancel()
		err := s.Server.Shutdown(ctx)
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) {
			_ = s.Server.Close()
			return err
		}
		return err
	})

	// Start server.
	tree.Go(func(ctx context.Context) error {
		err = s.Server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})

	return tree.Wait()
}

// Serve starts a HTTP and Connect gRPC server with sane defaults for FTL.
//
// Blocks until the context is cancelled.
func Serve(ctx context.Context, listen *url.URL, options ...Option) error {
	server, err := NewServer(ctx, listen, options...)
	if err != nil {
		return err
	}
	return server.Serve(ctx)
}
