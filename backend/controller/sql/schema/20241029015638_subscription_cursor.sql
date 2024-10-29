-- migrate:up

ALTER TABLE topic_subscriptions
DROP CONSTRAINT topic_subscriptions_cursor_fkey,
ADD CONSTRAINT topic_subscriptions_cursor_fkey
FOREIGN KEY (cursor)
REFERENCES topic_events (id)
ON DELETE RESTRICT;

-- migrate:down

