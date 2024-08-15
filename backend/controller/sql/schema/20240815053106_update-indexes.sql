-- migrate:up

DROP INDEX deployments_key_idx;
DROP INDEX artefacts_digest_idx;
DROP INDEX runners_key;
DROP INDEX ingress_requests_key_idx;
DROP INDEX cron_jobs_key_idx;
DROP INDEX topic_events_key_idx;
DROP INDEX topic_subscriptions_key_idx;

-- migrate:down

