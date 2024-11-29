-- migrate:up
ALTER TABLE deployments ADD COLUMN endpoint VARCHAR;

-- migrate:down

