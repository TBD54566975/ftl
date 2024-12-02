package server

import (
	"reflect"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/internal/schema"
)

func TopicHandle[E any, M ftl.TopicPartitionMap[E]](module, name string) reflection.VerbResource {
	handle := ftl.TopicHandle[E, M]{Ref: &schema.Ref{
		Name:   name,
		Module: module,
	}}
	return func() reflect.Value {
		return reflect.ValueOf(handle)
	}
}
