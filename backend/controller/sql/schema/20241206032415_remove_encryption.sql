-- migrate:up
ALTER TABLE async_calls DROP COLUMN request;
ALTER TABLE async_calls ADD COLUMN request json NOT NULL;
ALTER TABLE topic_events DROP COLUMN payload;
ALTER TABLE topic_events ADD COLUMN payload json NOT NULL;
DROP TABLE encryption_keys;
-- migrate:down

