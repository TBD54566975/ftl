-- migrate:up

ALTER TABLE topic_events
  ALTER COLUMN payload TYPE encrypted_async;

-- migrate:down
