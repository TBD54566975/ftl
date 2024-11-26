package schemaeventsource

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

func TestSchemaEventSource(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.TODO())
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	t.Cleanup(cancel)

	bind := must.Get(url.Parse("http://127.0.0.1:9090"))
	server := &mockSchemaService{changes: make(chan *ftlv1.PullSchemaResponse, 8)}
	go rpc.Serve(ctx, bind, rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, server)) //nolint:errcheck

	changes := New(ctx, rpc.Dial(ftlv1connect.NewSchemaServiceClient, bind.String(), log.Debug))

	send := func(t testing.TB, resp *ftlv1.PullSchemaResponse) {
		resp.ModuleName = resp.Schema.Name
		resp.DeploymentKey = proto.String(model.NewDeploymentKey(resp.ModuleName).String())
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())

		case server.changes <- resp:
		}
	}

	recv := func(t testing.TB) Event {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())

		case change := <-changes.Events():
			return change

		}
		panic("unreachable")
	}

	time1 := &schema.Module{
		Name: "time",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "time",
				Request:  &schema.Unit{},
				Response: &schema.Time{},
			},
		},
	}
	echo1 := &schema.Module{
		Name: "echo",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "echo",
				Request:  &schema.String{},
				Response: &schema.String{},
			},
		},
	}
	time2 := &schema.Module{
		Name: "time",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "time",
				Request:  &schema.Unit{},
				Response: &schema.Time{},
			},
			&schema.Verb{
				Name:     "timezone",
				Request:  &schema.Unit{},
				Response: &schema.String{},
			},
		},
	}

	t.Run("InitialSend", func(t *testing.T) {
		send(t, &ftlv1.PullSchemaResponse{
			More:       true,
			Schema:     (time1).ToProto().(*schemapb.Module), //nolint:forcetypeassert
			ChangeType: ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
		})

		send(t, &ftlv1.PullSchemaResponse{
			More:       false,
			Schema:     (echo1).ToProto().(*schemapb.Module), //nolint:forcetypeassert
			ChangeType: ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
		})

		var expected Event = EventUpsert{Module: time1, more: true}
		assertEqual(t, expected, recv(t))

		expected = EventUpsert{Module: echo1}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Modules: []*schema.Module{time1, echo1}}, changes.View())
		assertEqual(t, changes.View(), actual.Schema())
	})

	t.Run("Mutation", func(t *testing.T) {
		send(t, &ftlv1.PullSchemaResponse{
			More:       false,
			Schema:     (time2).ToProto().(*schemapb.Module), //nolint:forcetypeassert
			ChangeType: ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
		})

		var expected Event = EventUpsert{Module: time2}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Modules: []*schema.Module{time2, echo1}}, changes.View())
		assertEqual(t, changes.View(), actual.Schema())
	})

	// Verify that schemasync doesn't propagate "initial" again.
	t.Run("SimulatedReconnect", func(t *testing.T) {
		send(t, &ftlv1.PullSchemaResponse{
			More:       true,
			Schema:     (time2).ToProto().(*schemapb.Module), //nolint:forcetypeassert
			ChangeType: ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
		})
		send(t, &ftlv1.PullSchemaResponse{
			More:       false,
			Schema:     (echo1).ToProto().(*schemapb.Module), //nolint:forcetypeassert
			ChangeType: ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
		})

		var expected Event = EventUpsert{Module: time2}
		assertEqual(t, expected, recv(t))
		expected = EventUpsert{Module: echo1, more: false}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Modules: []*schema.Module{time2, echo1}}, changes.View())
		assertEqual(t, changes.View(), actual.Schema())
	})

	t.Run("Delete", func(t *testing.T) {
		send(t, &ftlv1.PullSchemaResponse{
			Schema:        (echo1).ToProto().(*schemapb.Module), //nolint:forcetypeassert
			ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED,
			ModuleRemoved: true,
		})
		var expected Event = EventRemove{Module: echo1, Deleted: true}
		actual := recv(t)
		assertEqual(t, expected, actual)
		assertEqual(t, &schema.Schema{Modules: []*schema.Module{time2}}, changes.View())
		assertEqual(t, changes.View(), actual.Schema())
	})
}

type mockSchemaService struct {
	ftlv1connect.UnimplementedSchemaServiceHandler
	changes chan *ftlv1.PullSchemaResponse
}

func (m *mockSchemaService) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], resp *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case change := <-m.changes:
			if err := resp.Send(change); err != nil {
				return fmt.Errorf("send change: %w", err)
			}
		}
	}
}

var _ ftlv1connect.SchemaServiceHandler = &mockSchemaService{}

func assertEqual[T comparable](t testing.TB, expected, actual T) {
	t.Helper()
	assert.Equal(t, expected, actual, assert.Exclude[optional.Option[model.DeploymentKey]](), assert.Exclude[*schema.Schema]())
}
