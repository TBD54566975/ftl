// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package sql

import (
	"context"

	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
)

const createOnlyIdentityKey = `-- name: CreateOnlyIdentityKey :exec
INSERT INTO identity_keys (id, private, public, verify_signature)
VALUES (1, $1, $2, $3)
`

func (q *Queries) CreateOnlyIdentityKey(ctx context.Context, private api.EncryptedIdentityColumn, public []byte, verifySignature []byte) error {
	_, err := q.db.ExecContext(ctx, createOnlyIdentityKey, private, public, verifySignature)
	return err
}

const getOnlyIdentityKey = `-- name: GetOnlyIdentityKey :one
SELECT private, public, verify_signature
FROM identity_keys
WHERE id = 1
`

type GetOnlyIdentityKeyRow struct {
	Private         api.EncryptedIdentityColumn
	Public          []byte
	VerifySignature []byte
}

func (q *Queries) GetOnlyIdentityKey(ctx context.Context) (GetOnlyIdentityKeyRow, error) {
	row := q.db.QueryRowContext(ctx, getOnlyIdentityKey)
	var i GetOnlyIdentityKeyRow
	err := row.Scan(&i.Private, &i.Public, &i.VerifySignature)
	return i, err
}
