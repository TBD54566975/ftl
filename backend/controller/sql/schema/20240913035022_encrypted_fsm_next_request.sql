-- migrate:up

ALTER TABLE fsm_next_event
    ALTER COLUMN request TYPE encrypted_async;

-- migrate:down
