package echo

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var db = ftl.PostgresDatabase("testdb")

type InsertRequest struct {
	Data string
}

type InsertResponse struct{}

//ftl:export
func Insert(ctx context.Context, req InsertRequest) (InsertResponse, error) {
	err := persistRequest(ctx, req)
	if err != nil {
		return InsertResponse{}, err
	}

	return InsertResponse{}, nil
}

func persistRequest(ctx context.Context, req InsertRequest) error {
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
