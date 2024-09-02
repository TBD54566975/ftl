-- migrate:up

DROP TRIGGER IF EXISTS deployments_notify_event ON deployments;
DROP TRIGGER IF EXISTS topics_notify_event ON topics;
DROP TRIGGER IF EXISTS topic_events_notify_event ON topic_events;
DROP FUNCTION IF EXISTS notify_event();

-- migrate:down

