package pgproxy_test

import (
	"context"
	"net"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/TBD54566975/ftl/internal/dev"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/pgproxy"
)

func TestPgProxy(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	client, proxy := net.Pipe()

	dsn, err := dev.SetupPostgres(ctx, "postgres:15.8", 0, false)
	assert.NoError(t, err)

	frontend := pgproto3.NewFrontend(client, client)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go pgproxy.HandleConnection(ctx, proxy, func(ctx context.Context, parameters map[string]string) (string, error) {
		return dsn, nil
	})

	t.Run("denies SSL", func(t *testing.T) {
		frontend.Send(&pgproto3.SSLRequest{})
		assert.NoError(t, frontend.Flush())

		assert.Equal(t, readOneByte(t, client), 'N')
	})

	t.Run("denies GSSEnc", func(t *testing.T) {
		frontend.Send(&pgproto3.GSSEncRequest{})
		assert.NoError(t, frontend.Flush())

		assert.Equal(t, readOneByte(t, client), 'N')
	})

	t.Run("authenticates with startup message", func(t *testing.T) {
		frontend.Send(&pgproto3.StartupMessage{ProtocolVersion: pgproto3.ProtocolVersionNumber, Parameters: map[string]string{
			"user": "ftl",
		}})
		assert.NoError(t, frontend.Flush())

		assertResponseType[*pgproto3.AuthenticationOk](t, frontend)
		for range 13 {
			assertResponseType[*pgproto3.ParameterStatus](t, frontend)
		}
		assertResponseType[*pgproto3.ReadyForQuery](t, frontend)
	})

	t.Run("proxies a query to the underlying DB", func(t *testing.T) {
		frontend.Send(&pgproto3.Query{String: "SELECT 1"})
		assert.NoError(t, frontend.Flush())

		assertResponseType[*pgproto3.RowDescription](t, frontend)
		assertResponseType[*pgproto3.DataRow](t, frontend)
		assertResponseType[*pgproto3.CommandComplete](t, frontend)
		assertResponseType[*pgproto3.ReadyForQuery](t, frontend)
	})
}

func readOneByte(t *testing.T, client net.Conn) byte {
	t.Helper()

	response := make([]byte, 1)
	n, err := client.Read(response)
	assert.NoError(t, err)
	assert.Equal(t, n, 1)
	return response[0]
}

func assertResponseType[T any](t *testing.T, f *pgproto3.Frontend) {
	t.Helper()

	var zero T
	resp, err := f.Receive()
	assert.NoError(t, err)
	_, ok := resp.(T)
	assert.True(t, ok, "expected response type %T, got %T", zero, resp)
}
