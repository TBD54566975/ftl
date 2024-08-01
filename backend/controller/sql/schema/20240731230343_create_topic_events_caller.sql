-- migrate:up

ALTER TABLE topic_events
    ADD COLUMN caller TEXT NOT NULL;

-- migrate:down
