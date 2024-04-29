package ftl

//TODO: should not be public in ftl package...

import (
	"context"
)

type CallOverrider interface {
	OverrideCall(ctx context.Context, callee Ref, req any) (override bool, resp any, err error)
}
type contextCallOverriderKey struct{}

func ApplyCallOverriderToContext(ctx context.Context, overrider CallOverrider) context.Context {
	return context.WithValue(ctx, contextCallOverriderKey{}, overrider)
}

func CallOverriderFromContext(ctx context.Context) (CallOverrider, bool) {
	if overrider, ok := ctx.Value(contextCallOverriderKey{}).(CallOverrider); ok {
		return overrider, true
	}
	return nil, false
}
