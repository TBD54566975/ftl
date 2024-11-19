package client

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/pubsub"
	"github.com/jpillora/backoff"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

// Change represents a change to a module.
//
//sumtype:decl
type Change interface {
	// If true there are more schema changes immediately following this one as part of the initial batch.
	// If false this is the last schema change in the initial batch, but others may follow later.
	IsInitialBatch() bool

	change()
}

// Upserted represents a module that was added or changed.
type Upserted struct {
	*schema.Module
	// If true there are more schema changes immediately following this one as part of the initial batch.
	// If false this is the last schema change in the initial batch, but others may follow later.
	InitialBatch bool
}

func (u Upserted) IsInitialBatch() bool { return u.InitialBatch }
func (Upserted) change()                {}

// Removed represents a module that was removed.
type Removed struct {
	*schema.Module
	// If true there are more schema changes immediately following this one as part of the initial batch.
	// If false this is the last schema change in the initial batch, but others may follow later.
	InitialBatch bool
}

func (r Removed) IsInitialBatch() bool { return r.InitialBatch }
func (Removed) change()                {}

type Client struct {
	*pubsub.Topic[Change]
	logger                     *log.Logger
	client                     SchemaServiceClient
	initialBatchReceived       atomic.Int32
	notifyInitialBatchReceived chan struct{}
}

type SchemaServiceClient interface {
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
	PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest]) (*connect.ServerStreamForClient[ftlv1.PullSchemaResponse], error)
}

// New creates a new client for the SchemaService that acts as a pubsub.Topic[Change].
func New(ctx context.Context, client SchemaServiceClient) *Client {
	topic := pubsub.New[Change]()
	return NewFromTopic(ctx, client, topic)
}

// NewFromTopic creates a new client from an existing pubsub.Topic[Change].
func NewFromTopic(ctx context.Context, client SchemaServiceClient, topic *pubsub.Topic[Change]) *Client {
	s := &Client{
		Topic:                      topic,
		logger:                     log.FromContext(ctx).Scope("schemaserviceclient"),
		client:                     client,
		initialBatchReceived:       atomic.NewInt32(0),
		notifyInitialBatchReceived: make(chan struct{}, 1),
	}
	s.logger.Debugf("Starting SchemaService client.")
	go rpc.RetryStreamingServerStream(ctx, "schemasource", backoff.Backoff{Max: time.Second * 2}, &ftlv1.PullSchemaRequest{}, client.PullSchema, s.update, rpc.AlwaysRetry())
	return s
}

// Ping the SchemaService.
func (s *Client) Ping(ctx context.Context) error {
	_, err := s.client.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if err != nil {
		return fmt.Errorf("SchemaService: %w", err)
	}
	return nil
}

// WaitForInitialSync blocks until the batch of events representing the initial full schema has been received.
func (s *Client) WaitForInitialSync(ctx context.Context) error {
	select {
	case <-s.notifyInitialBatchReceived:
		return nil

	case <-ctx.Done():
		return fmt.Errorf("wait for full schema sync: %w", ctx.Err())
	}
}

func (s *Client) update(ctx context.Context, resp *ftlv1.PullSchemaResponse) error {
	module, err := schema.ModuleFromProto(resp.Schema)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}
	// If the streaming connection has to be reestablished, we will receive the initial batch notification again, but
	// downstream consumers shouldn't have to care about this. TODO: Remove duplicate notifications.
	more := s.initialBatchReceived.Load() == 0 && resp.More
	switch resp.ChangeType {
	case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
		s.Topic.Publish(Upserted{module, more})
	default:
		s.Topic.Publish(Removed{module, more})
	}
	if s.initialBatchReceived.Load() == 0 && !resp.More {
		s.logger.Debugf("Received initial batch of schema changes.")
		s.initialBatchReceived.Store(1)
		close(s.notifyInitialBatchReceived)
	}
	return nil
}
