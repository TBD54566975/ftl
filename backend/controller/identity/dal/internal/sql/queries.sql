-- name: GetOnlyIdentityKey :one
SELECT private, public, verify_signature
FROM identity_keys
WHERE id = 1;

-- name: CreateOnlyIdentityKey :exec
INSERT INTO identity_keys (id, private, public, verify_signature)
VALUES (1, $1, $2, $3);
