-- name: GetOnlyEncryptionKey :one
SELECT key, verify_timeline, verify_async
FROM encryption_keys
WHERE id = 1;

-- name: CreateOnlyEncryptionKey :exec
INSERT INTO encryption_keys (id, key)
VALUES (1, $1);

-- name: UpdateEncryptionVerification :exec
UPDATE encryption_keys
SET verify_timeline = $1,
    verify_async = $2
WHERE id = 1;
