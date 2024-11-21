package pgproxy

import (
	"context"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/TBD54566975/ftl/internal/log"
)

type Config struct {
	Listen string `name:"listen" short:"l" help:"Address to listen on." env:"FTL_PROXY_PG_LISTEN" default:"127.0.0.1:5678"`
}

// PgProxy is a configurable proxy for PostgreSQL connections
type PgProxy struct {
	listenAddress      string
	connectionStringFn func(ctx context.Context, params map[string]string) (string, error)
}

// DSNConstructor is a function that constructs a new connection string from parameters of the incoming connection.
//
// parameters are pg connection parameters as described in https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS
type DSNConstructor func(ctx context.Context, params map[string]string) (string, error)

// New creates a new PgProxy.
//
// address is the address to listen on for incoming connections.
// connectionFn is a function that constructs a new connection string from parameters of the incoming connection.
func New(config Config, connectionFn DSNConstructor) *PgProxy {
	return &PgProxy{
		listenAddress:      config.Listen,
		connectionStringFn: connectionFn,
	}
}

// Start the proxy.
func (p *PgProxy) Start(ctx context.Context) error {
	logger := log.FromContext(ctx)

	listener, err := net.Listen("tcp", p.listenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", p.listenAddress, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Errorf(err, "failed to accept connection")
			continue
		}
		go HandleConnection(ctx, conn, p.connectionStringFn)
	}
}

// HandleConnection proxies a single connection.
//
// This should be run as the first thing after accepting a connection.
// It will block until the connection is closed.
func HandleConnection(ctx context.Context, conn net.Conn, connectionFn DSNConstructor) {
	defer conn.Close()

	logger := log.FromContext(ctx)
	logger.Infof("new connection established: %s", conn.RemoteAddr())

	backend, startup, err := connectBackend(ctx, conn)
	if err != nil {
		logger.Errorf(err, "failed to connect backend")
		return
	}
	logger.Debugf("startup message: %+v", startup)
	logger.Debugf("backend connected: %s", conn.RemoteAddr())

	dsn, err := connectionFn(ctx, startup.Parameters)
	if err != nil {
		handleBackendError(ctx, backend, err)
		return
	}

	frontend, err := connectFrontend(ctx, dsn)
	if err != nil {
		handleBackendError(ctx, backend, err)
		return
	}
	logger.Debugf("frontend connected")

	backend.Send(&pgproto3.AuthenticationOk{})
	backend.Send(&pgproto3.ReadyForQuery{})
	if err := backend.Flush(); err != nil {
		logger.Errorf(err, "failed to flush backend authentication ok")
		return
	}

	if err := proxy(ctx, backend, frontend); err != nil {
		logger.Warnf("disconnecting %s due to: %s", conn.RemoteAddr(), err)
		return
	}
	logger.Infof("terminating connection to %s", conn.RemoteAddr())
}

func handleBackendError(ctx context.Context, backend *pgproto3.Backend, err error) {
	logger := log.FromContext(ctx)
	logger.Errorf(err, "backend error")
	backend.Send(&pgproto3.ErrorResponse{
		Severity: "FATAL",
		Message:  err.Error(),
	})
	if err := backend.Flush(); err != nil {
		logger.Errorf(err, "failed to flush backend error response")
	}
}

// connectBackend establishes a connection according to https://www.postgresql.org/docs/current/protocol-flow.html
func connectBackend(_ context.Context, conn net.Conn) (*pgproto3.Backend, *pgproto3.StartupMessage, error) {
	backend := pgproto3.NewBackend(conn, conn)

	for {
		startup, err := backend.ReceiveStartupMessage()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to receive startup message: %w", err)
		}

		switch startup := startup.(type) {
		case *pgproto3.SSLRequest:
			// The client is requesting SSL connection. We don't support it.
			if _, err := conn.Write([]byte{'N'}); err != nil {
				return nil, nil, fmt.Errorf("failed to write ssl request response: %w", err)
			}
		case *pgproto3.CancelRequest:
			// TODO: implement cancel requests
			return backend, nil, nil
		case *pgproto3.StartupMessage:
			return backend, startup, nil
		case *pgproto3.GSSEncRequest:
			// The client is requesting GSS encryption. We don't support it.
			if _, err := conn.Write([]byte{'N'}); err != nil {
				return nil, nil, fmt.Errorf("failed to write gss encryption request response: %w", err)
			}
		default:
			return nil, nil, fmt.Errorf("unknown startup message: %T", startup)
		}
	}
}

func connectFrontend(ctx context.Context, dsn string) (*pgproto3.Frontend, error) {
	conn, err := pgconn.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to backend: %w", err)
	}
	frontend := pgproto3.NewFrontend(conn.Conn(), conn.Conn())

	return frontend, nil
}

func proxy(ctx context.Context, backend *pgproto3.Backend, frontend *pgproto3.Frontend) error {
	logger := log.FromContext(ctx)
	frontendMessages := make(chan pgproto3.BackendMessage)
	backendMessages := make(chan pgproto3.FrontendMessage)
	errors := make(chan error, 2)

	go func() {
		for {
			msg, err := backend.Receive()
			if err != nil {
				errors <- fmt.Errorf("failed to receive backend message: %w", err)
				return
			}
			logger.Tracef("backend message: %T", msg)
			backendMessages <- msg
		}
	}()

	go func() {
		for {
			msg, err := frontend.Receive()
			if err != nil {
				errors <- fmt.Errorf("failed to receive frontend message: %w", err)
				return
			}
			logger.Tracef("frontend message: %T", msg)
			frontendMessages <- msg
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done: %w", ctx.Err())
		case msg := <-backendMessages:
			frontend.Send(msg)
			if err := frontend.Flush(); err != nil {
				return fmt.Errorf("failed to flush frontend message: %w", err)
			}

			if _, ok := msg.(*pgproto3.Terminate); ok {
				return nil
			}
		case msg := <-frontendMessages:
			backend.Send(msg)
			if err := backend.Flush(); err != nil {
				return fmt.Errorf("failed to flush backend message: %w", err)
			}
		case err := <-errors:
			return err
		}
	}
}
