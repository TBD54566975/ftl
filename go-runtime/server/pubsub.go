package server

import (
	"reflect"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/go-runtime/ftl"
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
