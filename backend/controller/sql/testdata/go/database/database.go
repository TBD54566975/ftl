package database

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type MyDbConfig struct {
	ftl.DefaultPostgresDatabaseConfig
}

func (MyDbConfig) Name() string { return "testdb" }

type InsertRequest struct {
	Data string
}

type InsertResponse struct{}

//ftl:verb
func Insert(ctx context.Context, req InsertRequest, db ftl.DatabaseHandle[MyDbConfig]) (InsertResponse, error) {
	err := persistRequest(ctx, req, db)
	if err != nil {
		return InsertResponse{}, err
	}

	return InsertResponse{}, nil
}

func persistRequest(ctx context.Context, req InsertRequest, db ftl.DatabaseHandle[MyDbConfig]) error {
	_, err := db.Get(ctx).Exec(`CREATE TABLE IF NOT EXISTS requests
	       (
	         data TEXT,
	         created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
	         updated_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc')
	      );`)
	if err != nil {
		return err
	}
	_, err = db.Get(ctx).Exec("INSERT INTO requests (data) VALUES ($1);", req.Data)
	if err != nil {
		return err
	}
	return nil
}
