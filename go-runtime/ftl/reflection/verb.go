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
			args: resources,
			fn:   reflect.ValueOf(verb),
		}
		t.verbCalls[ref] = vi
	}
}

type VerbExec func(ctx context.Context, req any) (optional.Option[any], error)

type verbCall struct {
	args []reflect.Value
	fn   reflect.Value
}

func (v verbCall) Exec(ctx context.Context, req any) (optional.Option[any], error) {
	if v.fn.Kind() != reflect.Func {
		return optional.None[any](), fmt.Errorf("error invoking verb %v", v.fn)
	}

	args := make([]reflect.Value, 0, len(v.args)+2)
	args = append(args, reflect.ValueOf(ctx))
	args = append(args, reflect.ValueOf(req))
	args = append(args, v.args...)

	results := v.fn.Call(args)
	var resp optional.Option[any]
	var errValue reflect.Value
	if len(results) == 2 {
		resp = optional.Some(results[0].Interface())
		errValue = results[1]
	} else if len(results) == 1 {
		resp = optional.None[any]()
		errValue = results[0]
	} else {
		return optional.None[any](), fmt.Errorf("unexpected number of return values from verb %v", v.fn)
	}

	var err error
	if e := errValue.Interface(); e != nil {
		err = e.(error) //nolint:forcetypeassert
	}
	return resp, err
}
