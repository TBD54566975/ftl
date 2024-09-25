-- name: GetIdentityKeys :many
SELECT private, public, verify_signature
FROM identity_keys;

-- name: CreateOnlyIdentityKey :exec
INSERT INTO identity_keys (private, public, verify_signature)
VALUES ($1, $2, $3);
