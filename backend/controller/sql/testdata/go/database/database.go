package database

import (
	"context"
	"database/sql"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type MyDBConfig struct {
	ftl.DefaultPostgresDatabaseConfig
}

func (MyDBConfig) Name() string { return "testdb" }

type NewType struct {
	*sql.DB
}

type MyDBMapper struct {
	ftl.DatabaseHandle[MyDBConfig]
}

func (d MyDBMapper) Map(ctx context.Context, db *sql.DB) (NewType, error) {
	return NewType{db}, nil
}

type InsertRequest struct {
	Data string
}

type InsertResponse struct{}

//ftl:verb
func Mapped(ctx context.Context, req InsertRequest, db ftl.MappedHandle[MyDBMapper, *sql.DB, NewType]) (InsertResponse, error) {
	conn := db.Get(ctx)
	err := persistRequest(ctx, req, conn.DB)
	if err != nil {
		return InsertResponse{}, err
	}

	return InsertResponse{}, nil
}

//ftl:verb
func Insert(ctx context.Context, req InsertRequest, db ftl.DatabaseHandle[MyDBConfig]) (InsertResponse, error) {
	conn := db.Get(ctx)
	err := persistRequest(ctx, req, conn)
	if err != nil {
		return InsertResponse{}, err
	}

	return InsertResponse{}, nil
}

func persistRequest(ctx context.Context, req InsertRequest, db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS requests
	       (
	         data TEXT,
	         created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
	         updated_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc')
	      );`)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO requests (data) VALUES ($1);", req.Data)
	if err != nil {
		return err
	}
	return nil
}
