package controlplane

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/concurrency"
	"github.com/bufbuild/connect-go"
	"github.com/jpillora/backoff"
	"github.com/oklog/ulid/v2"

	"github.com/TBD54566975/ftl/controlplane/internal/dal"
	"github.com/TBD54566975/ftl/controlplane/internal/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

func TestControlPlaneRegisterRunnerHeartbeatClose(t *testing.T) {
	t.Skip("Sometimes failing with 'timed out waiting for server to bind'")
	db, client, bind, ctx := startForTesting(t)

	stream := client.RegisterRunner(ctx)
	t.Cleanup(func() { _, _ = stream.CloseAndReceive() })
	key := ulid.Make().String()
	err := stream.Send(&ftlv1.RegisterRunnerRequest{
		Key:      key,
		Language: "go",
		Endpoint: bind.String(),
	})
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)

	err = stream.Send(&ftlv1.RegisterRunnerRequest{
		Key:      key,
		Language: "go",
		Endpoint: bind.String(),
	})
	assert.NoError(t, err)

	eventually(t, func() bool {
		runners, err := db.GetIdleRunnersForLanguage(ctx, "go", 10)
		assert.NoError(t, err)
		return len(runners) > 0
	})

	runners, err := db.GetIdleRunnersForLanguage(ctx, "go", 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(runners))
	assert.Zero(t, runners[0].Deployment)
}

func TestControlPlaneRegisterRunnerHeartbeatTimeout(t *testing.T) {
	t.Skip("Sometimes failing with 'timed out waiting for server to bind'")
	db, client, bind, ctx := startForTesting(t)

	key := ulid.Make().String()
	stream := client.RegisterRunner(ctx)
	err := stream.Send(&ftlv1.RegisterRunnerRequest{
		Key:      key,
		Language: "go",
		Endpoint: bind.String(),
	})
	assert.NoError(t, err)

	eventually(t, func() bool {
		runners, err := db.GetIdleRunnersForLanguage(ctx, "go", 10)
		assert.NoError(t, err)
		return len(runners) > 0
	})
	time.Sleep(time.Second + time.Millisecond*200)
	_, err = stream.CloseAndReceive()
	assert.EqualError(t, err, "deadline_exceeded: heartbeat timeout")
	runners, err := db.GetIdleRunnersForLanguage(ctx, "go", 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(runners))
}

type combinedService struct {
	*Service
}

func (*combinedService) DeployToRunner(context.Context, *connect.Request[ftlv1.DeployToRunnerRequest]) (*connect.Response[ftlv1.DeployToRunnerResponse], error) {
	panic("unimplemented")
}

var _ ftlv1connect.RunnerServiceHandler = (*combinedService)(nil)

func startForTesting(t *testing.T) (*dal.DAL, ftlv1connect.ControlPlaneServiceClient, *url.URL, context.Context) {
	t.Helper()
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{Level: log.Warn}))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second) // Test timeout
	t.Cleanup(cancel)

	db := dal.New(sqltest.OpenForTesting(t))
	svc, err := New(ctx, db, 1*time.Second, 30*time.Second, 1024*1024)
	assert.NoError(t, err)

	combined := &combinedService{Service: svc}

	srv, err := rpc.NewServer(ctx, &url.URL{Scheme: "http", Host: "127.0.0.1:0"},
		rpc.GRPC(ftlv1connect.NewControlPlaneServiceHandler, combined),
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, combined),
		rpc.GRPC(ftlv1connect.NewRunnerServiceHandler, combined),
	)
	assert.NoError(t, err)
	ctx = concurrency.Call(ctx, func() error {
		return srv.Serve(ctx)
	})
	t.Cleanup(func() {
		err = srv.Server.Close()
		assert.NoError(t, err)
	})

	var bind *url.URL
	select {
	case bind = <-srv.Bind.Subscribe(make(chan *url.URL)):
		t.Logf("bound to %s", bind)

	case <-ctx.Done():
		t.Fatal("timed out waiting for server to bind")
	}

	// Create client and wait for server to become live.
	client := rpc.Dial(ftlv1connect.NewControlPlaneServiceClient, bind.String(), log.Error)

	// Wait for the server to come up.
	err = rpc.Wait(ctx, backoff.Backoff{Min: 100 * time.Millisecond, Max: 100 * time.Millisecond}, client)
	assert.NoError(t, err)

	return db, client, bind, ctx
}

func eventually(t testing.TB, f func() bool, msgandargs ...any) {
	t.Helper()
	b := &backoff.Backoff{
		Min:    10 * time.Millisecond,
		Max:    100 * time.Millisecond,
		Factor: 1.1,
	}
	for i := 0; i < 50; i++ {
		if f() {
			return
		}
		time.Sleep(b.Duration())
	}
	assert.False(t, false, msgandargs...)
}
