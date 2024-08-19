-- migrate:up

CREATE DOMAIN encrypted_timeline AS BYTEA;

ALTER TABLE timeline
    ALTER COLUMN payload TYPE encrypted_timeline;

CREATE DOMAIN encrypted_async AS BYTEA;

ALTER TABLE async_calls
    ALTER COLUMN request TYPE encrypted_async,
    ALTER COLUMN response TYPE encrypted_async;

-- migrate:down

