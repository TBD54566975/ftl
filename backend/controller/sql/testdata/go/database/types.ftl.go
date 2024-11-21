// Code generated by FTL. DO NOT EDIT.
package database

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type InsertClient func(context.Context, InsertRequest) (InsertResponse, error)

func init() {
	reflection.Register(
		reflection.Database[MyDbConfig]("testdb", server.InitPostgres),
		reflection.ProvideResourcesForVerb(
			Insert,
			server.DatabaseHandle[MyDbConfig]("postgres"),
		),
	)
}
