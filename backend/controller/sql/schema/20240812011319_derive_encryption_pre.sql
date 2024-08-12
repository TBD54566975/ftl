-- migrate:up

CREATE TABLE encryption_keys (
    id BIGINT PRIMARY KEY,
    encrypted BYTEA NOT NULL
);

ALTER TABLE events
    ADD COLUMN payload_new BYTEA,
    ADD COLUMN encryption_key_id BIGINT REFERENCES encryption_keys(id);
CREATE INDEX idx_events_encryption_key_id ON events(encryption_key_id);

ALTER TABLE async_calls
    ADD COLUMN request_new BYTEA,
    ADD COLUMN encryption_key_id BIGINT REFERENCES encryption_keys(id),
    DROP COLUMN response;
CREATE INDEX idx_async_calls_encryption_key_id ON async_calls(encryption_key_id);

ALTER TABLE topic_events
    ADD COLUMN payload_new BYTEA,
    ADD COLUMN encryption_key_id BIGINT REFERENCES encryption_keys(id);
CREATE INDEX idx_topic_events_encryption_key_id ON topic_events(encryption_key_id);

-- migrate:down

