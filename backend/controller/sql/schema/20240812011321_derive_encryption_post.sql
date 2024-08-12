-- migrate:up

-- ALTER TABLE events
--     DROP COLUMN payload,
--     RENAME COLUMN payload_new TO payload;
--
-- ALTER TABLE async_calls
--     DROP COLUMN request,
--     RENAME COLUMN request_new TO request;
--
-- ALTER TABLE topic_events
--     DROP COLUMN payload,
--     RENAME COLUMN payload_new TO payload;

-- migrate:down
