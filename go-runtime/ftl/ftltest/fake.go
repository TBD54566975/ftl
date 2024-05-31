package ftltest

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/TBD54566975/ftl/go-runtime/internal"
)

type fakeFTL struct {
	fsm           *fakeFSMManager
	mockMaps      map[uintptr]mapImpl
	allowMapCalls bool
}

// mapImpl is a function that takes an object and returns an object of a potentially different
// type but is not constrained by input/output type like ftl.Map.
type mapImpl func(context.Context) (any, error)

func newFakeFTL() *fakeFTL {
	return &fakeFTL{
		fsm:           newFakeFSMManager(),
		mockMaps:      make(map[uintptr]mapImpl),
		allowMapCalls: false,
	}
}

var _ internal.FTL = &fakeFTL{}

func (f *fakeFTL) FSMSend(ctx context.Context, fsm string, instance string, event any) error {
	return f.fsm.SendEvent(ctx, fsm, instance, event)
}

func (f *fakeFTL) PublishEvent(ctx context.Context, topic string, event any) error {
	panic("not implemented")
}

// addMapMock saves a new mock of ftl.Map to the internal map in fakeFTL.
//
// mockMap provides the whole mock implemention, so it gets called in place of both `fn`
// and `getter` in ftl.Map.
func (f *fakeFTL) addMapMock(mapper any, mockMap func(context.Context) (any, error)) {
	key := makeMapKey(mapper)
	f.mockMaps[key] = mockMap
}

func (f *fakeFTL) startAllowingMapCalls() {
	f.allowMapCalls = true
}

func (f *fakeFTL) CallMap(ctx context.Context, mapper any, mapImpl func(context.Context) (any, error)) any {
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
