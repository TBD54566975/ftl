-- migrate:up
ALTER TABLE deployments ADD COLUMN last_activated_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc');
UPDATE deployments SET last_activated_at = created_at;
-- migrate:down

