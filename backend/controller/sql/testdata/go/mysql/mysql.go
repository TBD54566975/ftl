package mysql

import (
	"context"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type MyDbConfig struct {
	ftl.DefaultMySQLDatabaseConfig
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

//ftl:verb
func Query(ctx context.Context, db ftl.DatabaseHandle[MyDbConfig]) (map[string]string, error) {
	var result string
	err := db.Get(ctx).QueryRowContext(ctx, "SELECT data FROM requests").Scan(&result)
	return map[string]string{"data": result}, err
}

func persistRequest(ctx context.Context, req InsertRequest, db ftl.DatabaseHandle[MyDbConfig]) error {
	_, err := db.Get(ctx).Exec("INSERT INTO requests (data) VALUES (?);", req.Data)
	if err != nil {
		return err
	}
	return nil
}
