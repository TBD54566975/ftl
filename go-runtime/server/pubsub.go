package server

import (
	"reflect"

	"github.com/TBD54566975/ftl/common/reflection"
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
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
