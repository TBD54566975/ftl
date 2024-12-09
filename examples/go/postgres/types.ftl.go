// Code generated by FTL. DO NOT EDIT.
package postgres

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type InsertClient func(context.Context, InsertRequest) (InsertResponse, error)

type QueryClient func(context.Context) ([]string, error)

func init() {
	reflection.Register(
		reflection.Database[MyDbConfig]("testdb", server.InitPostgres),
		reflection.ProvideResourcesForVerb(
			Insert,
			server.DatabaseHandle[MyDbConfig]("postgres"),
		),
		reflection.ProvideResourcesForVerb(
			Query,
			server.DatabaseHandle[MyDbConfig]("postgres"),
		),
	)
}
