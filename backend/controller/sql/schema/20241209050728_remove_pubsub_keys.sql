-- migrate:up
ALTER TABLE topic_subscribers DROP COLUMN deployment_id;
ALTER TABLE topic_subscribers ADD COLUMN deployment_key deployment_key NOT NULL;
ALTER TABLE topic_subscriptions DROP COLUMN deployment_id;
ALTER TABLE topic_subscriptions ADD COLUMN deployment_key deployment_key NOT NULL;
ALTER TABLE topic_subscriptions DROP COLUMN module_id;
ALTER TABLE topic_subscriptions ADD COLUMN module_name VARCHAR NOT NULL;
ALTER TABLE topics DROP COLUMN module_id;
ALTER TABLE topics ADD COLUMN module_name VARCHAR NOT NULL;

CREATE UNIQUE INDEX topic_subscriptions_module_name_idx ON topic_subscriptions(module_name, name);
CREATE UNIQUE INDEX topics_module_name_idx ON topics(module_name, name);
-- migrate:down

