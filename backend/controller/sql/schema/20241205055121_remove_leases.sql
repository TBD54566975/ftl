-- migrate:up
ALTER TABLE async_calls DROP COLUMN lease_id;
DROP TABLE leases;

-- migrate:down

