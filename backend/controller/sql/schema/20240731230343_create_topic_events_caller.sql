-- migrate:up

ALTER TABLE topic_events
    ADD COLUMN caller TEXT;

-- migrate:down
