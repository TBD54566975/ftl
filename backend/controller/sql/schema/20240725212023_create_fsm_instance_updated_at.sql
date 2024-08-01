-- migrate:up

ALTER TABLE fsm_instances
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc');

-- migrate:down