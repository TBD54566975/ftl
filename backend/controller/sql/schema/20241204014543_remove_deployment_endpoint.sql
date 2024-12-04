-- migrate:up

ALTER TABLE deployments DROP COLUMN endpoint;

-- migrate:down

