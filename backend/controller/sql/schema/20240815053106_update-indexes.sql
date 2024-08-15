-- migrate:up

DROP INDEX deployments_key_idx;
DROP INDEX artefacts_digest_idx;
DROP INDEX runners_key;
DROP INDEX ingress_requests_key_idx;
DROP INDEX cron_jobs_key_idx;
DROP INDEX topic_events_key_idx;
DROP INDEX topic_subscriptions_key_idx;

DROP INDEX IF EXISTS topics_module_name_idx;
ALTER TABLE topics
ADD CONSTRAINT topics_module_name_idx UNIQUE (module_id, name);

DROP INDEX IF EXISTS topic_subscriptions_module_name_idx;
ALTER TABLE topic_subscriptions
ADD CONSTRAINT topic_subscriptions_module_name_idx UNIQUE (module_id, name);

DROP INDEX IF EXISTS idx_fsm_instances_fsm_key;
ALTER TABLE fsm_instances
ADD CONSTRAINT idx_fsm_instances_fsm_key UNIQUE (fsm, key);

-- migrate:down

