-- migrate:up

ALTER TABLE async_calls
    ADD COLUMN cron_job_key cron_job_key;

ALTER TABLE async_calls
    ADD CONSTRAINT fk_async_calls_cron_job_key
    FOREIGN KEY (cron_job_key) REFERENCES cron_jobs(key)
    ON DELETE SET NULL;

CREATE INDEX idx_async_calls_cron_job_key
    ON async_calls (cron_job_key);

CREATE INDEX idx_async_calls_cron_job_key_scheduled_at
    ON async_calls (cron_job_key, scheduled_at);

ALTER TABLE cron_jobs
    DROP COLUMN state,
    ADD COLUMN last_execution TIMESTAMPTZ;

-- migrate:down

