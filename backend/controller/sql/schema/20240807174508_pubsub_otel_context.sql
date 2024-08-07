-- migrate:up

ALTER TABLE topic_events
    ADD COLUMN otel_context JSONB;

ALTER TABLE async_calls
    ADD COLUMN otel_context JSONB;

-- migrate:down

