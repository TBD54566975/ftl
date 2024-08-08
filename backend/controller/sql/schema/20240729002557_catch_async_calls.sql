-- migrate:up

ALTER TABLE async_calls
    ADD COLUMN catch_verb schema_ref, -- verb to call when retries have been exhausted
    ADD COLUMN catching BOOLEAN NOT NULL DEFAULT FALSE; -- whether the call needs to be caught

ALTER TABLE topic_subscribers
    ADD COLUMN catch_verb schema_ref;

-- migrate:down

