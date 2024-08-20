-- migrate:up

ALTER TABLE encryption_keys
    ADD COLUMN verify_timeline encrypted_timeline,
    ADD COLUMN verify_async encrypted_async;

-- migrate:down

ALTER TABLE encryption_keys
    DROP COLUMN verify_timeline,
    DROP COLUMN verify_async;
