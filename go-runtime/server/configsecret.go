package server

import (
	"reflect"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

func Config[T ftl.ConfigType](module, name string) reflection.VerbResource {
	cfg := ftl.Config[T]{Ref: reflection.Ref{Module: module, Name: name}}
	return func() reflect.Value {
		return reflect.ValueOf(cfg)
	}
}

func Secret[T ftl.SecretType](module, name string) reflection.VerbResource {
	secret := ftl.Secret[T]{Ref: reflection.Ref{Module: module, Name: name}}
	return func() reflect.Value {
		return reflect.ValueOf(secret)
	}
}
