package pgproxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/block/ftl/internal/log"
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
func New(listenAddress string, connectionFn DSNConstructor) *PgProxy {
	return &PgProxy{
		listenAddress:      listenAddress,
		connectionStringFn: connectionFn,
	}
}

type Started struct {
	Address *net.TCPAddr
}

// Start the proxy
func (p *PgProxy) Start(ctx context.Context, started chan<- Started) error {
	logger := log.FromContext(ctx)

	listener, err := net.Listen("tcp", p.listenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", p.listenAddress, err)
	}
	defer listener.Close()

	if started != nil {
		addr, ok := listener.Addr().(*net.TCPAddr)
		if !ok {
			panic("failed to get TCP address")
		}
		started <- Started{Address: addr}
	}

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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := log.FromContext(ctx)
	logger.Debugf("new connection established: %s", conn.RemoteAddr())

	backend, startup, err := connectBackend(ctx, conn)
	if err != nil {
		logger.Errorf(err, "failed to connect backend")
		return
	}
	if backend == nil {
		logger.Infof("client disconnected without startup message: %s", conn.RemoteAddr())
		return
	}
	logger.Tracef("startup message: %+v", startup)
	logger.Tracef("backend connected: %s", conn.RemoteAddr())

	hijacked, err := connectFrontend(ctx, connectionFn, startup)
	if err != nil {
		// try again, in case there was a credential rotation
		logger.Debugf("failed to connect frontend: %s, trying again", err)

		hijacked, err = connectFrontend(ctx, connectionFn, startup)
		if err != nil {
			handleBackendError(ctx, backend, err)
			return
		}
	}
	backend.Send(&pgproto3.AuthenticationOk{})
	logger.Debugf("frontend connected")
	for key, value := range hijacked.ParameterStatuses {
		backend.Send(&pgproto3.ParameterStatus{Name: key, Value: value})
	}

	backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
	if err := backend.Flush(); err != nil {
		logger.Errorf(err, "failed to flush backend authentication ok")
		return
	}

	if err := proxy(ctx, backend, hijacked.Frontend); err != nil {
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
func connectBackend(ctx context.Context, conn net.Conn) (*pgproto3.Backend, *pgproto3.StartupMessage, error) {
	logger := log.FromContext(ctx)

	backend := pgproto3.NewBackend(conn, conn)

	for {
		startup, err := backend.ReceiveStartupMessage()
		if errors.Is(err, io.EOF) {
			// some clients just terminate the connection and open a new one if it does not support SSL / GSS encryption
			return nil, nil, nil
		} else if err != nil {
			return nil, nil, fmt.Errorf("failed to receive startup message from %s: %w", conn.RemoteAddr(), err)
		}

		logger.Debugf("received startup message: %T from %s", startup, conn.RemoteAddr())

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

func connectFrontend(ctx context.Context, connectionFn DSNConstructor, startup *pgproto3.StartupMessage) (*pgconn.HijackedConn, error) {
	dsn, err := connectionFn(ctx, startup.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to construct dsn: %w", err)
	}

	conn, err := pgconn.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to backend: %w", err)
	}
	hijacked, err := conn.Hijack()
	if err != nil {
		return nil, fmt.Errorf("failed to hijack backend: %w", err)
	}
	return hijacked, nil
}

func proxy(ctx context.Context, backend *pgproto3.Backend, frontend *pgproto3.Frontend) error {
	logger := log.FromContext(ctx)
	errors := make(chan error, 2)

	go func() {
		for {
			msg, err := backend.Receive()
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err != nil {
				errors <- fmt.Errorf("failed to receive backend message: %w", err)
				return
			}
			logger.Tracef("backend message: %T", msg)
			frontend.Send(msg)
			err = frontend.Flush()
			if err != nil {
				errors <- fmt.Errorf("failed to receive backend message: %w", err)
				return
			}
			if _, ok := msg.(*pgproto3.Terminate); ok {
				return
			}
		}
	}()

	go func() {
		for {
			msg, err := frontend.Receive()
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err != nil {
				errors <- fmt.Errorf("failed to receive frontend message: %w", err)
				return
			}
			logger.Tracef("frontend message: %T", msg)
			backend.Send(msg)
			err = backend.Flush()
			if err != nil {
				errors <- fmt.Errorf("failed to receive backend message: %w", err)
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done: %w", ctx.Err())
		case err := <-errors:
			return err
		}
	}
}
