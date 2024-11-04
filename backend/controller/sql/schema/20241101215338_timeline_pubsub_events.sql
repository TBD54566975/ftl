-- migrate:up

ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'pubsub_publish';
ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'pubsub_consume';

-- migrate:down
