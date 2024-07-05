package ftltest

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/internal"
)

// pubSubEvent is a sum type for all events that can be published to the pubsub system.
// not to be confused with an event that gets published to a topic
//
//sumtype:decl
type pubSubEvent interface {
	// cronJobEvent is a marker to ensure that all events implement the interface.
	pubSubEvent()
}

// publishEvent holds an event to be published to a topic
type publishEvent struct {
	topic   *schema.Ref
	content any
}

func (publishEvent) pubSubEvent() {}

// subscriptionDidConsumeEvent indicates that a call to a subscriber has completed
type subscriptionDidConsumeEvent struct {
	subscription string
	err          error
}

func (subscriptionDidConsumeEvent) pubSubEvent() {}

type subscription struct {
	name        string
	topic       *schema.Ref
	cursor      optional.Option[int]
	isExecuting bool
	errors      map[int]error
}

type subscriber func(context.Context, any) error

type fakeFTL struct {
	t   testing.TB
	fsm *fakeFSMManager

	mockMaps      map[uintptr]mapImpl
	allowMapCalls bool
	configValues  map[string][]byte
	secretValues  map[string][]byte
	pubSub        *fakePubSub
}

// mapImpl is a function that takes an object and returns an object of a potentially different
// type but is not constrained by input/output type like ftl.Map.
type mapImpl func(context.Context) (any, error)

func newFakeFTL(ctx context.Context, t testing.TB) *fakeFTL {
	t.Helper()
	fake := &fakeFTL{
		t:             t,
		fsm:           newFakeFSMManager(),
		mockMaps:      map[uintptr]mapImpl{},
		allowMapCalls: false,
		configValues:  map[string][]byte{},
		secretValues:  map[string][]byte{},
		pubSub:        newFakePubSub(ctx, t),
	}

	return fake
}

var _ internal.FTL = &fakeFTL{}

func (f *fakeFTL) setConfig(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f.configValues[name] = data
	return nil
}

func (f *fakeFTL) GetConfig(ctx context.Context, name string, dest any) error {
	data, ok := f.configValues[name]
	if !ok {
		return fmt.Errorf("config value %q not found, did you remember to ctx := ftltest.Context(t, ftltest.WithDefaultProjectFile()) ?: %w", name, configuration.ErrNotFound)
	}
	return json.Unmarshal(data, dest)
}

func (f *fakeFTL) setSecret(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f.secretValues[name] = data
	return nil
}

func (f *fakeFTL) GetSecret(ctx context.Context, name string, dest any) error {
	data, ok := f.secretValues[name]
	if !ok {
		return fmt.Errorf("secret value %q not found, did you remember to ctx := ftltest.Context(t, ftltest.WithDefaultProjectFile()) ?: %w", name, configuration.ErrNotFound)
	}
	return json.Unmarshal(data, dest)
}

func (f *fakeFTL) FSMSend(ctx context.Context, fsm string, instance string, event any) error {
	return f.fsm.SendEvent(ctx, fsm, instance, event)
}

// addMapMock saves a new mock of ftl.Map to the internal map in fakeFTL.
//
// mockMap provides the whole mock implemention, so it gets called in place of both `fn`
// and `getter` in ftl.Map.
func addMapMock[T, U any](f *fakeFTL, mapper *ftl.MapHandle[T, U], mockMap func(context.Context) (U, error)) error {
	key, err := makeMapKey(mapper)
	if err != nil {
		return err
	}
	f.mockMaps[key] = func(ctx context.Context) (any, error) {
		return mockMap(ctx)
	}
	return nil
}

func (f *fakeFTL) startAllowingMapCalls() {
	f.allowMapCalls = true
}

func (f *fakeFTL) CallMap(ctx context.Context, mapper any, value any, mapImpl func(context.Context) (any, error)) any {
	f.t.Helper()
	key, err := makeMapKey(mapper)
	if err != nil {
		f.t.Fatalf("failed to call map: %v", err)
	}
	mockMap, ok := f.mockMaps[key]
	if ok {
		value, err := actuallyCallMap(ctx, mockMap)
		if err != nil {
			f.t.Fatalf("failed to call fake map: %v", err)
		}
		return value
	}
	if f.allowMapCalls {
		value, err := actuallyCallMap(ctx, mapImpl)
		if err != nil {
			f.t.Fatalf("failed to call map: %v", err)
		}
		return value
	}
	f.t.Fatalf("map calls not allowed in tests by default: ftltest.Context should be instantiated with either ftltest.WithMapsAllowed() or a mock for the specific map being called using ftltest.WhenMap(...)")
	return nil
}

func makeMapKey(mapper any) (uintptr, error) {
	v := reflect.ValueOf(mapper)
	if v.Kind() != reflect.Pointer {
		return 0, fmt.Errorf("fakeFTL received object that was not a pointer, expected *MapHandle")
	}
	underlying := v.Elem().Type().Name()
	if !strings.HasPrefix(underlying, "MapHandle[") {
		return 0, fmt.Errorf("fakeFTL received *%s, expected *MapHandle", underlying)
	}
	return v.Pointer(), nil
}

func actuallyCallMap(ctx context.Context, impl mapImpl) (any, error) {
	out, err := impl(ctx)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (f *fakeFTL) PublishEvent(ctx context.Context, topic *schema.Ref, event any) error {
	return f.pubSub.publishEvent(topic, event)
}
