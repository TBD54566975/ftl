package internal

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	"connectrpc.com/connect"
	"github.com/puzpuzpuz/xsync/v3"

	ftldeployment "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
	deploymentconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1/ftlv1connect"
	pubpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/publish/v1"
	pubconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/publish/v1/publishpbconnect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/internal/deploymentcontext"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

type mapCacheEntry struct {
	checksum [32]byte
	output   any
}

// RealFTL is the real implementation of the [internal.FTL] interface using the Controller.
type RealFTL struct {
	dmctx *deploymentcontext.DynamicDeploymentContext
	// Cache for Map() calls
	mapped *xsync.MapOf[uintptr, mapCacheEntry]
}

// New creates a new [RealFTL]
func New(dmctx *deploymentcontext.DynamicDeploymentContext) *RealFTL {
	return &RealFTL{
		dmctx:  dmctx,
		mapped: xsync.NewMapOf[uintptr, mapCacheEntry](),
	}
}

var _ FTL = &RealFTL{}

func (r *RealFTL) GetConfig(_ context.Context, name string, dest any) error {
	return r.dmctx.CurrentContext().GetConfig(name, dest)
}

func (r *RealFTL) GetSecret(_ context.Context, name string, dest any) error {
	return r.dmctx.CurrentContext().GetSecret(name, dest)
}

func (r *RealFTL) PublishEvent(ctx context.Context, topic *schema.Ref, event any, key string) error {
	caller := reflection.CallingVerb()
	if topic.Module != caller.Module {
		return fmt.Errorf("can not publish to another module's topic: %s", topic)
	}
	if err := publishToFTL(ctx, topic, event, caller); err != nil {
		return err
	}
	return publishToModule(ctx, topic, event, key, caller)
}

func publishToFTL(ctx context.Context, topic *schema.Ref, event any, caller schema.RefKey) error {
	// TODO: remove this once we have other pubsub moved over to kafka
	// For now we are publishing to both systems.
	client := rpc.ClientFromContext[deploymentconnect.DeploymentServiceClient](ctx)
	body, err := encoding.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event to controller: %w", err)
	}
	_, err = client.PublishEvent(ctx, connect.NewRequest(&ftldeployment.PublishEventRequest{
		Topic:  topic.ToProto().(*schemapb.Ref), //nolint: forcetypeassert
		Caller: caller.Name,
		Body:   body,
	}))
	if err != nil {
		return fmt.Errorf("failed to publish event to controller: %w", err)
	}
	return nil
}

func publishToModule(ctx context.Context, topic *schema.Ref, event any, key string, caller schema.RefKey) error {
	client := rpc.ClientFromContext[pubconnect.PublishServiceClient](ctx)
	body, err := encoding.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	_, err = client.PublishEvent(ctx, connect.NewRequest(&pubpb.PublishEventRequest{
		Topic:  topic.ToProto().(*schemapb.Ref), //nolint: forcetypeassert
		Caller: caller.Name,
		Body:   body,
		Key:    key,
	}))
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

func (r *RealFTL) CallMap(ctx context.Context, mapper any, value any, mapImpl func(context.Context) (any, error)) any {
	// Compute checksum of the input.
	inputData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal input data: %w", err)
	}
	checksum := sha256.Sum256(inputData)

	// Check cache.
	key := reflect.ValueOf(mapper).Pointer()
	cached, ok := r.mapped.Load(key)
	if ok && checksum == cached.checksum {
		return cached.output
	}

	// Map the value
	t, err := mapImpl(ctx)
	if err != nil {
		panic(err)
	}

	// Write the cache back.
	r.mapped.Store(key, mapCacheEntry{
		checksum: checksum,
		output:   t,
	})
	return t
}
