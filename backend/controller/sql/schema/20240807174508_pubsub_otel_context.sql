-- migrate:up

ALTER TABLE topic_events
    ADD COLUMN request_key TEXT,
    ADD COLUMN otel_context JSONB;

ALTER TABLE async_calls
    ADD COLUMN parent_request_key TEXT,
    ADD COLUMN otel_context JSONB;

ALTER TABLE events
    ADD COLUMN parent_request_id TEXT;

-- migrate:down

