-- migrate:up

ALTER TABLE cron_jobs
    DROP COLUMN state,
    ADD COLUMN last_execution TIMESTAMPTZ,
    ADD COLUMN last_async_call_id BIGINT;

ALTER TABLE cron_jobs
    ADD CONSTRAINT fk_cron_jobs_last_async_call_id
    FOREIGN KEY (last_async_call_id) REFERENCES async_calls(id)
    ON DELETE SET NULL;

DROP TYPE cron_job_state;

-- migrate:down

