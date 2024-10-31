-- migrate:up

ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'async_execute';

-- migrate:down
