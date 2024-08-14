-- migrate:up

ALTER TABLE events RENAME TO timeline;

-- migrate:down

