-- migrate:up

ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'cron_scheduled';

-- migrate:down

