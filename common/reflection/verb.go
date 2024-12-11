package reflection

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/types/optional"
)

// VerbResource is a function that registers a resource for a Verb.
type VerbResource func() reflect.Value

// ProvideResourcesForVerb registers any resources that must be provided when calling the given verb.
func ProvideResourcesForVerb(verb any, rs ...VerbResource) Registree {
	ref := FuncRef(verb)
	ref.Name = strings.TrimSuffix(ref.Name, "Client")
	return func(t *TypeRegistry) {
		resources := make([]reflect.Value, 0, len(rs))
		for _, r := range rs {
			resources = append(resources, r())
		}
		vi := verbCall{
			ref:  ref,
			args: resources,
			fn:   reflect.ValueOf(verb),
		}
		t.verbCalls[ref] = vi
	}
}

// VerbExec is a function for executing a verb.
type VerbExec func(ctx context.Context, req optional.Option[any]) (optional.Option[any], error)

type verbCall struct {
	ref  Ref
	args []reflect.Value
	fn   reflect.Value
}

// Exec executes the verb with the given context and request, adding any resources that were provided.
func (v verbCall) Exec(ctx context.Context, req optional.Option[any]) (optional.Option[any], error) {
	if v.fn.Kind() != reflect.Func {
		return optional.None[any](), fmt.Errorf("error invoking verb %v", v.fn)
	}

	var args []reflect.Value
	args = append(args, reflect.ValueOf(ctx))
	if r, ok := req.Get(); ok {
		args = append(args, reflect.ValueOf(r))
	}

	tryCall := func(args []reflect.Value) (results []reflect.Value, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		results = v.fn.Call(args)
		return results, err
	}

	results, err := tryCall(append(args, v.args...))
	if err != nil {
		return optional.None[any](), err
	}

	var resp optional.Option[any]
	var errValue reflect.Value
	switch len(results) {
	case 0:
		return optional.None[any](), nil
	case 1:
		resp = optional.None[any]()
		errValue = results[0]
	case 2:
		resp = optional.Some(results[0].Interface())
		errValue = results[1]
	default:
		return optional.None[any](), fmt.Errorf("unexpected number of return values from verb %s", v.ref)
	}
	var fnError error
	if e := errValue.Interface(); e != nil {
		fnError = e.(error) //nolint:forcetypeassert
	}
	return resp, fnError
}
