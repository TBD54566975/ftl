-- migrate:up

ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'ingress';

-- migrate:down

