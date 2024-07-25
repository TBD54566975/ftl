package internal

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	"connectrpc.com/connect"
	"github.com/puzpuzpuz/xsync/v3"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/internal/modulecontext"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type mapCacheEntry struct {
	checksum [32]byte
	output   any
}

// RealFTL is the real implementation of the [internal.FTL] interface using the Controller.
type RealFTL struct {
	dmctx *modulecontext.DynamicModuleContext
	// Cache for Map() calls
	mapped *xsync.MapOf[uintptr, mapCacheEntry]
}

// New creates a new [RealFTL]
func New(dmctx *modulecontext.DynamicModuleContext) *RealFTL {
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

func (r *RealFTL) FSMSend(ctx context.Context, fsm, instance string, event any) error {
	client := rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
	body, err := encoding.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	_, err = client.SendFSMEvent(ctx, connect.NewRequest(&ftlv1.SendFSMEventRequest{
		Fsm:      &schemapb.Ref{Module: reflection.Module(), Name: fsm},
		Instance: instance,
		Event:    schema.TypeToProto(reflection.ReflectTypeToSchemaType(reflect.TypeOf(event))),
		Body:     body,
	}))
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}
	return nil
}

func (r *RealFTL) PublishEvent(ctx context.Context, topic *schema.Ref, event any) error {
	if topic.Module != reflection.Module() {
		return fmt.Errorf("can not publish to another module's topic: %s", topic)
	}
	client := rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
	body, err := encoding.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	_, err = client.PublishEvent(ctx, connect.NewRequest(&ftlv1.PublishEventRequest{
		Topic: topic.ToProto().(*schemapb.Ref), //nolint: forcetypeassert
		Body:  body,
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
