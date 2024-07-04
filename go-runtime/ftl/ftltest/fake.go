package ftltest

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/alecthomas/types/optional"
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

func newFakeFTL(ctx context.Context) *fakeFTL {
	fake := &fakeFTL{
		fsm:           newFakeFSMManager(),
		mockMaps:      map[uintptr]mapImpl{},
		allowMapCalls: false,
		configValues:  map[string][]byte{},
		secretValues:  map[string][]byte{},
		pubSub:        newFakePubSub(ctx),
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
		return fmt.Errorf("secret value %q not found. Did you remember to ctx := ftltest.Context(ftltest.WithDefaultProjectFile()) ?: %w", name, configuration.ErrNotFound)
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
		return fmt.Errorf("config value %q not found. Did you remember to ctx := ftltest.Context(ftltest.WithDefaultProjectFile()) ?: %w", name, configuration.ErrNotFound)
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
func addMapMock[T, U any](f *fakeFTL, mapper *ftl.MapHandle[T, U], mockMap func(context.Context) (U, error)) {
	key := makeMapKey(mapper)
	f.mockMaps[key] = func(ctx context.Context) (any, error) {
		return mockMap(ctx)
	}
}

func (f *fakeFTL) startAllowingMapCalls() {
	f.allowMapCalls = true
}

func (f *fakeFTL) CallMap(ctx context.Context, mapper any, value any, mapImpl func(context.Context) (any, error)) any {
	key := makeMapKey(mapper)
	mockMap, ok := f.mockMaps[key]
	if ok {
		return actuallyCallMap(ctx, mockMap)
	}
	if f.allowMapCalls {
		return actuallyCallMap(ctx, mapImpl)
	}
	panic("map calls not allowed in tests by default. ftltest.Context should be instantiated with either ftltest.WithMapsAllowed() or a mock for the specific map being called using ftltest.WhenMap(...)")
}

func makeMapKey(mapper any) uintptr {
	v := reflect.ValueOf(mapper)
	if v.Kind() != reflect.Pointer {
		panic("fakeFTL received object that was not a pointer, expected *MapHandle")
	}
	underlying := v.Elem().Type().Name()
	if !strings.HasPrefix(underlying, "MapHandle[") {
		panic(fmt.Sprintf("fakeFTL received *%s, expected *MapHandle", underlying))
	}
	return v.Pointer()
}

func actuallyCallMap(ctx context.Context, impl mapImpl) any {
	out, err := impl(ctx)
	if err != nil {
		panic(err)
	}
	return out
}

func (f *fakeFTL) PublishEvent(ctx context.Context, topic *schema.Ref, event any) error {
	return f.pubSub.publishEvent(topic, event)
}
