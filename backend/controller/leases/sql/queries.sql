-- name: NewLease :one
INSERT INTO leases (
  idempotency_key,
  key,
  expires_at,
  metadata
)
VALUES (
  gen_random_uuid(),
  @key::lease_key,
  (NOW() AT TIME ZONE 'utc') + @ttl::interval,
  sqlc.narg('metadata')::JSONB
)
RETURNING idempotency_key;

-- name: RenewLease :one
UPDATE leases
SET expires_at = (NOW() AT TIME ZONE 'utc') + @ttl::interval
WHERE idempotency_key = @idempotency_key AND key = @key::lease_key
RETURNING true;

-- name: ReleaseLease :one
DELETE FROM leases
WHERE idempotency_key = @idempotency_key AND key = @key::lease_key
RETURNING true;

-- name: ExpireLeases :one
WITH expired AS (
    DELETE FROM leases
    WHERE expires_at < NOW() AT TIME ZONE 'utc'
    RETURNING 1
)
SELECT COUNT(*)
FROM expired;

-- name: GetLeaseInfo :one
SELECT expires_at, metadata FROM leases WHERE key = @key::lease_key;
