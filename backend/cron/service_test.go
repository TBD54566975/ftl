package cron

import (
	"context"
	"net/url"
	"os"
	"sort"
	"testing"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/timeline"
	schemapb "github.com/TBD54566975/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/routing"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

type verbClient struct {
	requests chan *ftlv1.CallRequest
}

var _ routing.CallClient = (*verbClient)(nil)

func (v *verbClient) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	v.requests <- req.Msg
	return connect.NewResponse(&ftlv1.CallResponse{Response: &ftlv1.CallResponse_Body{Body: []byte("{}")}}), nil
}

func TestCron(t *testing.T) {
	eventSource := schemaeventsource.NewUnattached()
	module := &schema.Module{
		Name: "echo",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "echo",
				Request:  &schema.Unit{},
				Response: &schema.Unit{},
				Metadata: []schema.Metadata{
					&schema.MetadataCronJob{Cron: "*/2 * * * * *"},
				},
			},
			&schema.Verb{
				Name:     "time",
				Request:  &schema.Unit{},
				Response: &schema.Unit{},
				Metadata: []schema.Metadata{
					&schema.MetadataCronJob{Cron: "*/2 * * * * *"},
				},
			},
		},
	}
	eventSource.Publish(schemaeventsource.EventUpsert{
		Deployment: optional.Some(model.NewDeploymentKey("echo")),
		Module:     module,
	})

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{Level: log.Trace}))
	timelineEndpoint, err := url.Parse("http://localhost:8080")
	assert.NoError(t, err)

	timelineClient := timeline.NewClient(ctx, timelineEndpoint)
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	t.Cleanup(cancel)

	wg, ctx := errgroup.WithContext(ctx)

	requestsch := make(chan *ftlv1.CallRequest, 8)
	client := &verbClient{
		requests: requestsch,
	}

	wg.Go(func() error { return Start(ctx, eventSource, client, timelineClient) })

	requests := make([]*ftlv1.CallRequest, 0, 2)

done:
	for range 2 {
		select {
		case <-ctx.Done():
			t.Fatalf("timed out: %s", ctx.Err())

		case request := <-requestsch:
			requests = append(requests, request)
			if len(requests) == 2 {
				break done
			}
		}
	}

	cancel()

	sort.SliceStable(requests, func(i, j int) bool {
		return requests[i].Verb.Name < requests[j].Verb.Name
	})
	assert.Equal(t, []*ftlv1.CallRequest{
		{
			Metadata: &ftlv1.Metadata{},
			Verb:     &schemapb.Ref{Module: "echo", Name: "echo"},
			Body:     []byte("{}"),
		},
		{
			Metadata: &ftlv1.Metadata{},
			Verb:     &schemapb.Ref{Module: "echo", Name: "time"},
			Body:     []byte("{}"),
		},
	}, requests, assert.Exclude[*schemapb.Position]())

	err = wg.Wait()
	assert.IsError(t, err, context.Canceled)
}
