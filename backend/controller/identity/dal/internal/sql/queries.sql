-- name: GetOnlyIdentityKey :one
SELECT key, verify_signature
FROM identity_keys
WHERE id = 1;

-- name: CreateOnlyIdentityKey :exec
INSERT INTO identity_keys (id, key)
VALUES (1, $1);

-- name: UpdateIdentityVerification :exec
UPDATE identity_keys
SET verify_signature = $1
WHERE id = 1;
