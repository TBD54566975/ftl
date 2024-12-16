package database

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/go-runtime/ftl/ftltest"
)

func TestDatabase(t *testing.T) {
	ctx := ftltest.Context(
		ftltest.WithCallsAllowedWithinModule(),
		ftltest.WithDatabase[MyDbConfig](),
	)

	_, err := ftltest.Call[InsertClient, InsertRequest, InsertResponse](ctx, InsertRequest{Data: "unit test 1"})
	assert.NoError(t, err)
	list, err := getAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, "unit test 1", list[0])

	ctx = ftltest.Context(
		ftltest.WithCallsAllowedWithinModule(),
		ftltest.WithDatabase[MyDbConfig](),
	)

	_, err = ftltest.Call[InsertClient, InsertRequest, InsertResponse](ctx, InsertRequest{Data: "unit test 2"})
	assert.NoError(t, err)
	list, err = getAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, "unit test 2", list[0])
}

func TestOptionOrdering(t *testing.T) {
	ctx := ftltest.Context(
		ftltest.WithCallsAllowedWithinModule(),
		ftltest.WithDatabase[MyDbConfig](), // <--- consumes DSNs
	)

	_, err := ftltest.Call[InsertClient, InsertRequest, InsertResponse](ctx, InsertRequest{Data: "unit test 1"})
	assert.NoError(t, err)
	list, err := getAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, "unit test 1", list[0])
}

func getAll(ctx context.Context) ([]string, error) {
	db, err := ftltest.GetDatabaseHandle[MyDbConfig]()
	if err != nil {
		return nil, err
	}
	rows, err := db.Get(ctx).Query("SELECT data FROM requests ORDER BY created_at;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []string{}
	for rows.Next() {
		var data string
		err := rows.Scan(&data)
		if err != nil {
			return nil, err
		}
		list = append(list, data)
	}
	return list, nil
}
