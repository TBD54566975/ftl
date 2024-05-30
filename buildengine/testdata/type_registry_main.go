// Code generated by FTL. DO NOT EDIT.
package main

import (
	"context"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"

	"ftl/other"
)

func init() {
	reflection.Register(
		reflection.SumType[other.SecondTypeEnum](
			*new(other.A),
			*new(other.B),
		),
		reflection.SumType[other.TypeEnum](
			*new(other.MyBool),
			*new(other.MyBytes),
			*new(other.MyFloat),
			*new(other.MyInt),
			*new(other.MyTime),
			*new(other.MyList),
			*new(other.MyMap),
			*new(other.MyString),
			*new(other.MyStruct),
			*new(other.MyOption),
			*new(other.MyUnit),
		),
	)
}

func main() {
	verbConstructor := server.NewUserVerbServer("other",
		server.HandleCall(other.Echo),
	)
	plugin.Start(context.Background(), "other", verbConstructor, ftlv1connect.VerbServiceName, ftlv1connect.NewVerbServiceHandler)
}
